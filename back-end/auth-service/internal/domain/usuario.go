package domain

import "time"

// RolUsuario define los roles posibles del sistema
type RolUsuario string

const (
	RolAdmin        RolUsuario = "admin"
	RolOperario     RolUsuario = "operario"
	RolSupervisor   RolUsuario = "supervisor"
	RolVisualizador RolUsuario = "visualizador"
)

// Usuario representa la entidad de dominio del usuario (sin contraseña)
type Usuario struct {
	IDUsuario          int32      `json:"id_usuario"`
	Empleado           string     `json:"empleado"`
	NombreCompleto     string     `json:"nombre_completo"`
	DescripcionCargo   string     `json:"descripcion_cargo"`
	Email              string     `json:"email"`
	Rol                RolUsuario `json:"rol"`
	Activo             bool       `json:"activo"`
	FechaCreacion      time.Time  `json:"fecha_creacion"`
	FechaActualizacion time.Time  `json:"fecha_actualizacion"`
}

// PeticionLogin contiene las credenciales enviadas por el cliente
type PeticionLogin struct {
	Email      string `json:"email"`
	Contrasena string `json:"contrasena"`
}

// PeticionLoginCedula contiene la cédula para login de empleados de vw_Ubicaciones
type PeticionLoginCedula struct {
	Cedula string `json:"cedula"`
}

// RespuestaLogin contiene el token JWT, los datos del usuario y la ruta de redirección
type RespuestaLogin struct {
	Token          string  `json:"token"`
	Usuario        Usuario `json:"usuario"`
	RedireccionarA string  `json:"redireccionarA"`
}

// ReclaimsJWT son los datos embebidos dentro del token JWT
type ReclaimsJWT struct {
	IDUsuario int32      `json:"id_usuario"`
	Email     string     `json:"email"`
	Rol       RolUsuario `json:"rol"`
}

// PeticionCrearUsuario contiene los datos para registrar un usuario
type PeticionCrearUsuario struct {
	Empleado         string `json:"empleado"`
	NombreCompleto   string `json:"nombre_completo"`
	DescripcionCargo string `json:"descripcion_cargo"`
	Email            string `json:"email"`
	Contrasena       string `json:"contrasena"`
	Rol              string `json:"rol"`
}

// RespuestaCrearUsuario contiene el resultado de registrar un usuario
type RespuestaCrearUsuario struct {
	Mensaje string  `json:"mensaje"`
	Usuario Usuario `json:"usuario"`
}
