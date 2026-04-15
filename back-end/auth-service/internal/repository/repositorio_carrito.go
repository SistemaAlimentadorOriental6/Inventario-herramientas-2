package repository

import (
	"auth-service/internal/domain"
	"context"
)

// RepositorioCarrito define el acceso a asignaciones_carritos en MySQL
type RepositorioCarrito interface {
	ObtenerCarritosPorUsuario(ctx context.Context, idUsuario int32) ([]int32, error)
	ObtenerUsuariosConCarritos(ctx context.Context) ([]domain.UsuarioConCarritos, error)
	ExisteUsuarioOperarioActivo(ctx context.Context, idUsuario int32) (bool, error)
	AsignarCarritoAUsuario(ctx context.Context, idUsuario int32, numeroCarrito int32) ([]int32, error)
	QuitarCarritoDeUsuario(ctx context.Context, idUsuario int32, numeroCarrito int32) (bool, error)
}

// RepositorioUbicacion define el acceso a vw_Ubicaciones en SQL Server
type RepositorioUbicacion interface {
	ObtenerFilasUbicaciones(ctx context.Context) ([]domain.FilaUbicacion, error)
	ObtenerUsuariosUnicos(ctx context.Context) ([]domain.UsuarioUbicacion, error)
	ObtenerPorCedula(ctx context.Context, cedula string) (*domain.UsuarioUbicacion, error)
}
