package dbresolver

import (
	"context"
	"database/sql"
	"errors"

	"gorm.io/gorm"
)

var _ gorm.ConnPool = &EmptyConnPool{}

var ErrEmptyConnPool = errors.New("empty connection pool, please health check the backend server")

type EmptyConnPool struct{}

func (EmptyConnPool) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) { //nolint:revive
	return nil, ErrEmptyConnPool
}

func (EmptyConnPool) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) { //nolint:revive
	return nil, ErrEmptyConnPool
}

func (EmptyConnPool) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) { //nolint:revive
	return nil, ErrEmptyConnPool
}

func (EmptyConnPool) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row { //nolint:revive
	return nil
}
