package repository

import (
	"auth-service/internal/domain"
	"context"
)

// RepositorioUsuario define el contrato de acceso a datos para usuarios
type RepositorioUsuario interface {
	ObtenerPorEmail(ctx context.Context, email string) (*domain.Usuario, string, error)
	ObtenerPorId(ctx context.Context, id int32) (*domain.Usuario, error)
	CrearUsuario(ctx context.Context, peticion domain.PeticionCrearUsuario) (*domain.Usuario, error)
}
