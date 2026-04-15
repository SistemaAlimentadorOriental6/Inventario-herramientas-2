package repository

import (
	"auth-service/internal/domain"
	"context"
)

// RepositorioInventario define las operaciones de persistencia para el inventario
type RepositorioInventario interface {
	GuardarRegistros(ctx context.Context, registros []domain.RegistroInventario) (int, error)

	// ObtenerReferenciasGuardadas retorna un set de referencias ya guardadas para usuario+carrito
	ObtenerReferenciasGuardadas(ctx context.Context, idUsuario int32, numCarrito int32) (map[string]bool, error)

	// ContarCompletadosPorUsuario retorna un mapa de numero_carrito → cantidad completada para un usuario
	ContarCompletadosPorUsuario(ctx context.Context, idUsuario int32) (map[int32]int, error)
}
