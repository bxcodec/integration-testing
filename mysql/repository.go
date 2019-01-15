package mysql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/bxcodec/integration-testing/models"
	"github.com/go-sql-driver/mysql"
)

// MysqlHandler ...
type MysqlHandler struct {
	DB *sql.DB
}

func NewHandler(db *sql.DB) *MysqlHandler {
	return &MysqlHandler{
		DB: db,
	}
}

const (
	// MysqlDuplicateStatusCode  (Read for mysql documentation about this error code)
	MysqlDuplicateStatusCode = 1062
)

// Store ...
func (m MysqlHandler) Store(ctx context.Context, c *models.Category) error {
	now := time.Now()
	query := `INSERT category SET name=?, slug=?, created_at=?, updated_at=?`
	res, err := m.DB.ExecContext(ctx, query, c.Name, c.Slug, now, now)
	if err != nil {
		errMysal, ok := err.(*mysql.MySQLError)
		if !ok {
			return err
		}

		if errMysal.Number == MysqlDuplicateStatusCode {
			return fmt.Errorf("Category is Duplicated")
		}
	}

	lastID, err := res.LastInsertId()
	if err != nil {
		return err
	}
	c.ID = lastID
	return nil
}

// GetByID ...
func (m MysqlHandler) GetByID(ctx context.Context, id int64) (models.Category, error) {
	res := models.Category{}
	query := `SELECT id, name, slug, created_at, updated_at FROM category WHERE id=?`

	row := m.DB.QueryRowContext(ctx, query, id)
	err := row.Scan(
		&res.ID,
		&res.Name,
		&res.Slug,
		&res.CreatedAt,
		&res.UpdatedAt,
	)
	return res, err
}

// Fetch ...
func (m MysqlHandler) Fetch(ctx context.Context, param models.Filter) ([]models.Category, error) {
	res := []models.Category{}

	queryBuilder := squirrel.Select("id", "name", "slug", "created_at", "updated_at").From("category")
	queryBuilder = queryBuilder.OrderBy("id DESC")
	if param.Cursor != "" {
		queryBuilder = queryBuilder.Where(squirrel.Lt{"id": param.Cursor})
	}
	if param.Keyword != "" {
		queryBuilder = queryBuilder.Where("name LIKE ?", fmt.Sprint("%", param.Keyword, "%"))
	}
	if param.Num > 0 {
		queryBuilder = queryBuilder.Limit(uint64(param.Num))
	}

	query, params, err := queryBuilder.ToSql()
	if err != nil {
		return res, err
	}

	rows, err := m.DB.QueryContext(ctx, query, params...)
	if err != nil {
		return res, err
	}
	for rows.Next() {
		cat := models.Category{}
		err := rows.Scan(
			&cat.ID,
			&cat.Name,
			&cat.Slug,
			&cat.CreatedAt,
			&cat.UpdatedAt,
		)
		if err != nil {
			return res, err
		}
		res = append(res, cat)

	}
	if rows.Err() != nil {
		return res, rows.Err()
	}
	return res, nil
}

// GetBySlug ...
func (m MysqlHandler) GetBySlug(ctx context.Context, slug string) (models.Category, error) {
	res := models.Category{}
	query := `SELECT id, name, slug, created_at, updated_at FROM category WHERE slug=?`
	row := m.DB.QueryRowContext(ctx, query, slug)
	err := row.Scan(
		&res.ID,
		&res.Name,
		&res.Slug,
		&res.CreatedAt,
		&res.UpdatedAt,
	)
	return res, err
}

// Update ...
func (m MysqlHandler) Update(ctx context.Context, c *models.Category) error {
	query := `UPDATE category SET name=?, updated_at=? WHERE id=?`
	res, err := m.DB.ExecContext(ctx, query, c.Name, time.Now(), c.ID)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("Notting Affected")
	}
	return nil
}

// Delete ...
func (m MysqlHandler) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM category  WHERE id=?`
	res, err := m.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return fmt.Errorf("Notting Affected")
	}
	return nil
}
