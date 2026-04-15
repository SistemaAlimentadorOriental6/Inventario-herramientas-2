// Código generado por SQLC - NO EDITAR MANUALMENTE
package sqlc

import (
	"database/sql"
	"time"
)

// RolUsuario representa el enum del campo rol
type RolUsuario string

const (
	RolAdmin      RolUsuario = "admin"
	RolOperario   RolUsuario = "operario"
	RolSupervisor RolUsuario = "supervisor"
)

// Usuario mapea la tabla `usuarios` de MySQL
type Usuario struct {
	IDUsuario          int32          `json:"id_usuario"`
	Empleado           sql.NullString `json:"empleado"`
	NombreCompleto     sql.NullString `json:"nombre_completo"`
	DescripcionCargo   sql.NullString `json:"descripcion_cargo"`
	Email              sql.NullString `json:"email"`
	Contrasena         sql.NullString `json:"contrasena"`
	Rol                RolUsuario     `json:"rol"`
	Activo             int8           `json:"activo"`
	FechaCreacion      time.Time      `json:"fecha_creacion"`
	FechaActualizacion time.Time      `json:"fecha_actualizacion"`
}
