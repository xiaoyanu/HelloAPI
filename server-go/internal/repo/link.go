package repo

import (
	"context"
	"github.com/zxz/helloapi/internal/model"
)

func (s *Store) GetLinkList(ctx context.Context, pageSize int, offset int, desc bool) ([]model.Link, error) {
	const query = "select * from helloapi_links ORDER BY CASE WHEN ? THEN created END DESC, CASE WHEN NOT ? THEN created END" +
		" LIMIT ? OFFSET ?"
	return selectRows[model.Link](ctx, s.query, query, desc, desc, pageSize, offset)
}

func (s *Store) GetLinkListCount(ctx context.Context) (int64, error) {
	const query = "select count(*) from helloapi_links"
	return scalar[int64](ctx, s.query, query)
}

func (s *Store) GetLinkListSearch(ctx context.Context, keyword string, pageSize int, offset int) ([]model.Link, error) {
	const query = "SELECT * FROM `helloapi_links` WHERE (? IS NULL OR ? = '' OR `name` LIKE CONCAT('%', ?, '%') OR `link_id`" +
		" LIKE CONCAT('%', ?, '%') OR `desc` = ?) ORDER BY `created` DESC LIMIT ? OFFSET ?"
	return selectRows[model.Link](ctx, s.query, query, keyword, keyword, keyword, keyword, keyword, pageSize, offset)
}

func (s *Store) GetLinkListSearchCount(ctx context.Context, keyword string) (int64, error) {
	const query = "SELECT COUNT(*) FROM `helloapi_links` WHERE (? IS NULL OR ? = '' OR `name` LIKE CONCAT('%', ?, '%') OR" +
		" `link_id` LIKE CONCAT('%', ?, '%') OR `desc` = ?)"
	return scalar[int64](ctx, s.query, query, keyword, keyword, keyword, keyword, keyword)
}

func (s *Store) InsertLink(ctx context.Context, link *model.Link) (int64, error) {
	const query = "INSERT INTO `helloapi_links` (`name`, `url`, `desc`, `avatar`) VALUES (?, ?, ?, ?)"
	result, err := s.query.ExecContext(ctx, query, link.Name, link.URL, link.Desc, link.Avatar)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	link.ID = int(id)
	return result.RowsAffected()
}

func (s *Store) UpdateLink(ctx context.Context, link *model.Link) (int64, error) {
	const query = "UPDATE `helloapi_links` SET `name` = ?, `url` = ?, `desc` = ?, `avatar` = ? WHERE `link_id` = ?"
	return affected(ctx, s.query, query, link.Name, link.URL, link.Desc, link.Avatar, link.ID)
}

func (s *Store) DeleteLink(ctx context.Context, linkID int) (int64, error) {
	const query = "DELETE FROM `helloapi_links` WHERE `link_id` = ?"
	return affected(ctx, s.query, query, linkID)
}

func (s *Store) GetLinkByID(ctx context.Context, linkID int) (*model.Link, error) {
	const query = "SELECT * FROM `helloapi_links` WHERE `link_id` = ?"
	return getRow[model.Link](ctx, s.query, query, linkID)
}
