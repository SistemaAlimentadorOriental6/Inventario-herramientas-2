package service

import (
	"auth-service/internal/domain"
	"auth-service/internal/repository"
	"context"
)

// ServicioPrestamo define el contrato del servicio de préstamo
type ServicioPrestamo interface {
	ObtenerItemsPrestamo(ctx context.Context) (*domain.RespuestaItemsPrestamo, error)
}

type servicioPrestamoImpl struct {
	repoUNOEE repository.RepositorioUNOEE
}

// NuevoServicioPrestamo crea la instancia del servicio
func NuevoServicioPrestamo(repoUNOEE repository.RepositorioUNOEE) ServicioPrestamo {
	return &servicioPrestamoImpl{
		repoUNOEE: repoUNOEE,
	}
}

// ObtenerItemsPrestamo retorna todos los ítems disponibles para préstamo
func (s *servicioPrestamoImpl) ObtenerItemsPrestamo(ctx context.Context) (*domain.RespuestaItemsPrestamo, error) {
	items, err := s.repoUNOEE.ObtenerItemsPrestamo(ctx)
	if err != nil {
		return nil, err
	}

	return &domain.RespuestaItemsPrestamo{
		Total: len(items),
		Items: items,
	}, nil
}
