// Código generado por SQLC - NO EDITAR MANUALMENTE
package sqlc

import (
	"context"
	"database/sql"
)

	// Querier define la interfaz de todas las consultas disponibles
type Querier interface {
	ObtenerUsuarioPorEmail(ctx context.Context, email string) (Usuario, error)
	ObtenerUsuarioPorId(ctx context.Context, id int32) (Usuario, error)
}

type Queries struct {
	db DBTX
}

type DBTX interface {
	ExecContext(context.Context, string, ...interface{}) (sql.Result, error)
	PrepareContext(context.Context, string) (*sql.Stmt, error)
	QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
}

func New(db DBTX) *Queries {
	return &Queries{db: db}
}
