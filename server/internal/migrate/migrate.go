package migrate

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"hash/crc32"
	"log/slog"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
)

//go:embed sql/*.sql
var migrations embed.FS

const historyDDL = "CREATE TABLE flyway_schema_history (" +
	"installed_rank INT NOT NULL PRIMARY KEY, version VARCHAR(50), description VARCHAR(200) NOT NULL, " +
	"type VARCHAR(20) NOT NULL, script VARCHAR(1000) NOT NULL, checksum INT, installed_by VARCHAR(100) NOT NULL, " +
	"installed_on TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP, execution_time INT NOT NULL, success TINYINT(1) NOT NULL, " +
	"INDEX flyway_schema_history_s_idx (success)) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4"

const insertHistory = "INSERT INTO flyway_schema_history " +
	"(installed_rank, version, description, type, script, checksum, installed_by, execution_time, success) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)"

var migrationName = regexp.MustCompile(`^V([0-9]+(?:\.[0-9]+)*)__([^/]+)\.sql$`)

type script struct {
	version, description, name string
	body                       string
	checksum                   int32
}

func checksum(raw []byte) int32 {
	text := strings.TrimPrefix(string(raw), "\ufeff")
	// Flyway 使用 Java BufferedReader，逐行读取时移除 CR、LF 和 CRLF。
	text = strings.NewReplacer("\r", "", "\n", "").Replace(text)
	return int32(crc32.ChecksumIEEE([]byte(text)))
}

func normalizedVersion(version string) []string {
	parts := strings.Split(version, ".")
	for i, part := range parts {
		parts[i] = strings.TrimLeft(part, "0")
	}
	for len(parts) > 0 && parts[len(parts)-1] == "" {
		parts = parts[:len(parts)-1]
	}
	return parts
}

func compareVersion(first, second string) int {
	a, b := normalizedVersion(first), normalizedVersion(second)
	for i := range min(len(a), len(b)) {
		if len(a[i]) != len(b[i]) {
			return len(a[i]) - len(b[i])
		}
		if compared := strings.Compare(a[i], b[i]); compared != 0 {
			return compared
		}
	}
	return len(a) - len(b)
}

func embeddedScripts() ([]script, error) {
	files, err := migrations.ReadDir("sql")
	if err != nil {
		return nil, err
	}
	scripts := make([]script, 0, len(files))
	for _, file := range files {
		parts := migrationName.FindStringSubmatch(file.Name())
		if parts == nil {
			return nil, fmt.Errorf("迁移文件名无效：%s", file.Name())
		}
		raw, err := migrations.ReadFile("sql/" + file.Name())
		if err != nil {
			return nil, err
		}
		scripts = append(scripts, script{
			version: parts[1], description: strings.ReplaceAll(parts[2], "_", " "), name: file.Name(),
			body: strings.TrimPrefix(string(raw), "\ufeff"), checksum: checksum(raw),
		})
	}
	slices.SortFunc(scripts, func(a, b script) int { return compareVersion(a.version, b.version) })
	for i := 1; i < len(scripts); i++ {
		if compareVersion(scripts[i-1].version, scripts[i].version) == 0 {
			return nil, fmt.Errorf("迁移版本重复：%s 与 %s", scripts[i-1].name, scripts[i].name)
		}
	}
	return scripts, nil
}

type historyEntry struct {
	kind     string
	checksum sql.NullInt64
}

func Run(ctx context.Context, cfg *mysql.Config, logger *slog.Logger) (err error) {
	scripts, err := embeddedScripts()
	if err != nil {
		return err
	}
	migrationConfig := cfg.Clone()
	migrationConfig.MultiStatements = true
	connector, err := mysql.NewConnector(migrationConfig)
	if err != nil {
		return err
	}
	db := sql.OpenDB(connector)
	defer db.Close()
	conn, err := db.Conn(ctx)
	if err != nil {
		return fmt.Errorf("连接迁移数据库: %w", err)
	}
	defer conn.Close()
	var locked sql.NullInt64
	if err := conn.QueryRowContext(ctx, "SELECT GET_LOCK('helloapi_migrate', 60)").Scan(&locked); err != nil {
		return fmt.Errorf("获取迁移锁: %w", err)
	}
	if !locked.Valid || locked.Int64 != 1 {
		return errors.New("未获取数据库迁移锁 helloapi_migrate")
	}
	defer func() {
		releaseCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if _, releaseErr := conn.ExecContext(releaseCtx, "SELECT RELEASE_LOCK('helloapi_migrate')"); releaseErr != nil && err == nil {
			err = fmt.Errorf("释放迁移锁: %w", releaseErr)
		}
	}()

	var installedBy string
	if err := conn.QueryRowContext(ctx, "SELECT SUBSTRING_INDEX(CURRENT_USER(), '@', 1)").Scan(&installedBy); err != nil {
		return err
	}
	if err := initializeHistory(ctx, conn, installedBy, logger); err != nil {
		return err
	}
	entries, rank, baseline, err := readHistory(ctx, conn)
	if err != nil {
		return err
	}

	for _, migration := range scripts {
		if err := ctx.Err(); err != nil {
			return err
		}
		key := versionKey(migration.version)
		if entry, ok := entries[key]; ok {
			if entry.kind == "SQL" && (!entry.checksum.Valid || entry.checksum.Int64 != int64(migration.checksum)) {
				logger.Warn("已执行的迁移校验和不一致", "版本", migration.version, "文件", migration.name,
					"历史校验和", entry.checksum, "脚本校验和", migration.checksum)
			}
			continue
		}
		if baseline != "" && compareVersion(migration.version, baseline) <= 0 {
			continue
		}

		start := time.Now()
		// 一旦开始执行，保持持锁连接直到写入历史记录。
		migrationCtx := context.WithoutCancel(ctx)
		_, migrationErr := conn.ExecContext(migrationCtx, migration.body)
		// MySQL DDL 可能部分提交。执行失败时记录失败状态，阻止下次启动继续变更。
		rank++
		elapsed := min(time.Since(start).Milliseconds(), int64(1<<31-1))
		if migrationErr != nil {
			_, recordErr := conn.ExecContext(migrationCtx, insertHistory, rank, migration.version, migration.description, "SQL",
				migration.name, migration.checksum, installedBy, elapsed, false)
			cause := fmt.Errorf("迁移 %s 失败，请检查数据库及 Flyway 历史: %w", migration.name, migrationErr)
			if recordErr != nil {
				return errors.Join(cause, fmt.Errorf("记录迁移失败状态: %w", recordErr))
			}
			return cause
		}
		if _, err := conn.ExecContext(migrationCtx, insertHistory, rank, migration.version, migration.description, "SQL",
			migration.name, migration.checksum, installedBy, elapsed, true); err != nil {
			return fmt.Errorf("记录已执行迁移 %s: %w", migration.name, err)
		}
		logger.Info("已执行数据库迁移", "版本", migration.version, "文件", migration.name)
	}
	return nil
}

func initializeHistory(ctx context.Context, conn *sql.Conn, installedBy string, logger *slog.Logger) error {
	var tables, histories int
	if err := conn.QueryRowContext(ctx, "SELECT COUNT(*), COUNT(CASE WHEN table_name = 'flyway_schema_history' THEN 1 END) FROM information_schema.tables WHERE table_schema = DATABASE()").Scan(&tables, &histories); err != nil {
		return fmt.Errorf("检查数据库迁移历史: %w", err)
	}
	if histories > 0 {
		return nil
	}
	if _, err := conn.ExecContext(ctx, historyDDL); err != nil {
		return fmt.Errorf("创建 Flyway 历史表: %w", err)
	}
	if tables == 0 {
		return nil
	}

	const marker = "<< Flyway Baseline >>"
	if _, err := conn.ExecContext(ctx, insertHistory, 1, "1", marker, "BASELINE", marker, nil, installedBy, 0, true); err != nil {
		return fmt.Errorf("写入 Flyway 基线: %w", err)
	}
	logger.Info("已为现有数据库建立 Flyway 基线", "版本", "1")
	return nil
}

func readHistory(ctx context.Context, conn *sql.Conn) (map[string]historyEntry, int, string, error) {
	rows, err := conn.QueryContext(ctx, "SELECT installed_rank, version, type, script, checksum, success FROM flyway_schema_history ORDER BY installed_rank")
	if err != nil {
		return nil, 0, "", fmt.Errorf("读取 Flyway 历史: %w", err)
	}
	defer rows.Close()
	entries := make(map[string]historyEntry)
	var rank int
	var baseline string
	for rows.Next() {
		var installedRank int
		var version sql.NullString
		var name string
		var entry historyEntry
		var success bool
		if err := rows.Scan(&installedRank, &version, &entry.kind, &name, &entry.checksum, &success); err != nil {
			return nil, 0, "", err
		}
		if !success {
			return nil, 0, "", fmt.Errorf("Flyway 历史存在失败迁移（序号 %s，脚本 %s），请修复后重试", strconv.Itoa(installedRank), name)
		}
		rank = max(rank, installedRank)
		if !version.Valid {
			continue
		}
		entries[versionKey(version.String)] = entry
		if entry.kind == "BASELINE" {
			baseline = version.String
		}
	}
	return entries, rank, baseline, rows.Err()
}

func versionKey(version string) string { return strings.Join(normalizedVersion(version), ".") }
