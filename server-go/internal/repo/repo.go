package repo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type Store struct {
	db    *sqlx.DB
	query sqlx.ExtContext
}

func New(db *sqlx.DB) *Store {
	// 与 MyBatis 行为一致：结果集中多出的列直接忽略，避免后续迁移加列后 select * 扫描报错。
	// 由该句柄开启的事务会继承此设置。
	db = db.Unsafe()
	return &Store{db: db, query: db}
}

func (s *Store) inTx(ctx context.Context, fn func(*Store) error) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := fn(&Store{db: s.db, query: tx}); err != nil {
		return err
	}
	return tx.Commit()
}

func scalar[T any](ctx context.Context, queryer sqlx.QueryerContext, query string, args ...any) (T, error) {
	var value T
	err := sqlx.GetContext(ctx, queryer, &value, query, args...)
	return value, err
}

func getRow[T any](ctx context.Context, queryer sqlx.QueryerContext, query string, args ...any) (*T, error) {
	value, err := scalar[T](ctx, queryer, query, args...)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &value, nil
}

func selectRows[T any](ctx context.Context, queryer sqlx.QueryerContext, query string, args ...any) ([]T, error) {
	list := make([]T, 0)
	err := sqlx.SelectContext(ctx, queryer, &list, query, args...)
	return list, err
}

func unique[T any](ctx context.Context, queryer sqlx.QueryerContext, query string, args ...any) (*T, error) {
	values, err := selectRows[T](ctx, queryer, query+" LIMIT 2", args...)
	if err != nil || len(values) == 0 {
		return nil, err
	}
	if len(values) > 1 {
		return nil, fmt.Errorf("应返回一条记录，实际返回了多条记录")
	}
	return &values[0], nil
}

func execute(ctx context.Context, db sqlx.ExecerContext, query string, args ...any) error {
	_, err := db.ExecContext(ctx, query, args...)
	return err
}

func affected(ctx context.Context, db sqlx.ExecerContext, query string, args ...any) (int64, error) {
	result, err := db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
