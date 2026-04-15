// Código generado por SQLC - NO EDITAR MANUALMENTE
package sqlc

import (
	"context"
)

const obtenerUsuarioPorEmail = `-- name: ObtenerUsuarioPorEmail :one
SELECT
  id_usuario, empleado, nombre_completo, descripcion_cargo,
  email, contrasena, rol, activo, fecha_creacion, fecha_actualizacion
FROM usuarios
WHERE email = ? AND activo = 1
LIMIT 1`

// ObtenerUsuarioPorEmail busca un usuario activo por su email
func (q *Queries) ObtenerUsuarioPorEmail(ctx context.Context, email string) (Usuario, error) {
	row := q.db.QueryRowContext(ctx, obtenerUsuarioPorEmail, email)
	var u Usuario
	err := row.Scan(
		&u.IDUsuario,
		&u.Empleado,
		&u.NombreCompleto,
		&u.DescripcionCargo,
		&u.Email,
		&u.Contrasena,
		&u.Rol,
		&u.Activo,
		&u.FechaCreacion,
		&u.FechaActualizacion,
	)
	return u, err
}

const obtenerUsuarioPorId = `-- name: ObtenerUsuarioPorId :one
SELECT
  id_usuario, empleado, nombre_completo, descripcion_cargo,
  email, contrasena, rol, activo, fecha_creacion, fecha_actualizacion
FROM usuarios
WHERE id_usuario = ? AND activo = 1
LIMIT 1`

// ObtenerUsuarioPorId busca un usuario activo por su ID
func (q *Queries) ObtenerUsuarioPorId(ctx context.Context, id int32) (Usuario, error) {
	row := q.db.QueryRowContext(ctx, obtenerUsuarioPorId, id)
	var u Usuario
	err := row.Scan(
		&u.IDUsuario,
		&u.Empleado,
		&u.NombreCompleto,
		&u.DescripcionCargo,
		&u.Email,
		&u.Contrasena,
		&u.Rol,
		&u.Activo,
		&u.FechaCreacion,
		&u.FechaActualizacion,
	)
	return u, err
}
