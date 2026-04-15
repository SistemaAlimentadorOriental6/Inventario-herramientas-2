package repository

import (
	"auth-service/internal/domain"
	"context"
)

// RepositorioUNOEE define el contrato de acceso a datos para la base UNOEE
type RepositorioUNOEE interface {
	ObtenerItemsPrestamo(ctx context.Context) ([]domain.ItemPrestamo, error)
	ObtenerExistenciaPorReferencia(ctx context.Context, referencia string) (float64, error)
}
