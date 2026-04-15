package service

import (
	"auth-service/internal/domain"
	"auth-service/internal/repository"
	"context"
)

type ServicioInventario interface {
	ProcesarGuardado(ctx context.Context, registros []domain.RegistroInventario) (*domain.RespuestaGuardado, error)
}

type servicioInventarioImpl struct {
	repo repository.RepositorioInventario
}

func NuevoServicioInventario(repo repository.RepositorioInventario) ServicioInventario {
	return &servicioInventarioImpl{repo: repo}
}

func (s *servicioInventarioImpl) ProcesarGuardado(ctx context.Context, registros []domain.RegistroInventario) (*domain.RespuestaGuardado, error) {
	procesados, err := s.repo.GuardarRegistros(ctx, registros)
	if err != nil {
		return nil, err
	}

	return &domain.RespuestaGuardado{
		Mensaje:    "Inventario guardado exitosamente",
		Procesados: procesados,
		Registros:  registros, // Retorna los registros guardados
	}, nil
}
