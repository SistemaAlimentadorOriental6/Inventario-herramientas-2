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
	repoUNOEE    repository.RepositorioUNOEE
	repoAdmon    repository.RepositorioAdmon
	repoPrestamo repository.RepositorioPrestamo
}

// NuevoServicioPrestamo crea la instancia del servicio
func NuevoServicioPrestamo(repoUNOEE repository.RepositorioUNOEE, repoAdmon repository.RepositorioAdmon, repoPrestamo repository.RepositorioPrestamo) ServicioPrestamo {
	return &servicioPrestamoImpl{
		repoUNOEE:    repoUNOEE,
		repoAdmon:    repoAdmon,
		repoPrestamo: repoPrestamo,
	}
}

// ObtenerItemsPrestamo retorna todos los ítems disponibles para préstamo
func (s *servicioPrestamoImpl) ObtenerItemsPrestamo(ctx context.Context) (*domain.RespuestaItemsPrestamo, error) {
	items, err := s.repoUNOEE.ObtenerItemsPrestamo(ctx)
	if err != nil {
		return nil, err
	}

	// Extraer todas las referencias únicas para consultar en ADMON
	referenciasMap := make(map[string]bool)
	var referencias []string
	for _, item := range items {
		if item.Referencia != "" && !referenciasMap[item.Referencia] {
			referenciasMap[item.Referencia] = true
			referencias = append(referencias, item.Referencia)
		}
	}

	// Consultar nombres inteligentes en ADMON
	nombresAdmon, err := s.repoAdmon.ObtenerNombresPorReferencia(ctx, referencias)
	if err != nil {
		// Logueamos el error pero no bloqueamos el proceso principal
	}

	// Mapear los nombres de ADMON a los items y descontar prestamos activos de MySQL
	for i := range items {
		ref := items[i].Referencia
		if nombresAdmon != nil {
			if nombre, ok := nombresAdmon[ref]; ok {
				items[i].NombreInteligente = nombre
			}
		}

		if s.repoPrestamo != nil && ref != "" {
			prestados, err := s.repoPrestamo.SumarCantidadPrestadaPorReferencia(ctx, ref)
			if err == nil && prestados > 0 {
				items[i].Existencia = items[i].Existencia - prestados
				if items[i].Existencia < 0 {
					items[i].Existencia = 0
				}
			}
		}
	}

	return &domain.RespuestaItemsPrestamo{
		Total: len(items),
		Items: items,
	}, nil
}
