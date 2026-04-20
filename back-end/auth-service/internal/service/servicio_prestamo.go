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
	repoAdmon repository.RepositorioAdmon
}

// NuevoServicioPrestamo crea la instancia del servicio
func NuevoServicioPrestamo(repoUNOEE repository.RepositorioUNOEE, repoAdmon repository.RepositorioAdmon) ServicioPrestamo {
	return &servicioPrestamoImpl{
		repoUNOEE: repoUNOEE,
		repoAdmon: repoAdmon,
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
		// ya que el nombre de UNOEE es el fallback
		// (Aquí idealmente usaríamos un logger real)
	}

	// Mapear los nombres de ADMON a los items
	if nombresAdmon != nil {
		for i := range items {
			if nombre, ok := nombresAdmon[items[i].Referencia]; ok {
				items[i].NombreInteligente = nombre
			}
		}
	}

	return &domain.RespuestaItemsPrestamo{
		Total: len(items),
		Items: items,
	}, nil
}
