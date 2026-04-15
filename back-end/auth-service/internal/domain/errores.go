package domain

import "errors"

// Errores de dominio del microservicio de autenticación
var (
	ErrCredencialesInvalidas = errors.New("email o contraseña incorrectos")
	ErrUsuarioInactivo       = errors.New("la cuenta de usuario está inactiva")
	ErrUsuarioNoEncontrado   = errors.New("usuario no encontrado")
	ErrUsuarioExistente      = errors.New("ya existe un usuario con ese email o número de empleado")
	ErrTokenInvalido         = errors.New("token de autenticación inválido o expirado")
)
