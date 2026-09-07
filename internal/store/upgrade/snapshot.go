package upgrade

import (
	"context"
	"database/sql"

	"github.com/jackc/pgx/v5"
)

type SQLSnapshotQueries interface {
	SQLSnapshot
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}
type PGSnapshotQueries interface {
	PGSnapshot
	Query(context.Context, string, ...any) (pgx.Rows, error)
}

type migrationRows interface {
	Next() bool
	Scan(...any) error
	Err() error
}
