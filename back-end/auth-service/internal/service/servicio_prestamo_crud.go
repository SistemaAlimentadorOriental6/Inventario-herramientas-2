package service

import (
	"auth-service/internal/domain"
	"auth-service/internal/repository"
	"auth-service/internal/utils"
	"context"
	"fmt"
	"strings"
	"time"
)

// ServicioPrestamoCRUD define el contrato del servicio de préstamos
type ServicioPrestamoCRUD interface {
	CrearPrestamo(ctx context.Context, req domain.CrearPrestamoRequest, idUsuario int32, nombreUsuario string) (*domain.RespuestaPrestamo, error)
	ObtenerPrestamo(ctx context.Context, id int64) (*domain.Prestamo, error)
	ListarPrestamos(ctx context.Context, estado, cedula, referencia string) (*domain.ListadoPrestamosResponse, error)
	ListarPrestamosPorOperario(ctx context.Context, cedula string, estado string) (*domain.ListadoPrestamosResponse, error)
	DevolverPrestamo(ctx context.Context, id int64, req domain.DevolverPrestamoRequest, idUsuario int32) error
	CalcularStockDisponible(ctx context.Context, referencia string, existenciaUNOEE float64) (*domain.StockDisponibleResponse, error)
}

type servicioPrestamoCRUDImpl struct {
	repoPrestamo repository.RepositorioPrestamo
	repoUNOEE    repository.RepositorioUNOEE
}

// NuevoServicioPrestamoCRUD crea el servicio de préstamos
func NuevoServicioPrestamoCRUD(repoPrestamo repository.RepositorioPrestamo, repoUNOEE repository.RepositorioUNOEE) ServicioPrestamoCRUD {
	return &servicioPrestamoCRUDImpl{
		repoPrestamo: repoPrestamo,
		repoUNOEE:    repoUNOEE,
	}
}

// CrearPrestamo crea un nuevo préstamo validando stock disponible
func (s *servicioPrestamoCRUDImpl) CrearPrestamo(ctx context.Context, req domain.CrearPrestamoRequest, idUsuario int32, nombreUsuario string) (*domain.RespuestaPrestamo, error) {
	// Normalizar referencia (quitar espacios sobrantes)
	req.Referencia = strings.TrimSpace(req.Referencia)

	// Validar stock disponible
	existencia, err := s.repoUNOEE.ObtenerExistenciaPorReferencia(ctx, req.Referencia)
	if err != nil {
		return nil, fmt.Errorf("error al consultar existencia: %w", err)
	}

	prestados, err := s.repoPrestamo.SumarCantidadPrestadaPorReferencia(ctx, req.Referencia)
	if err != nil {
		return nil, fmt.Errorf("error al consultar préstamos activos: %w", err)
	}

	disponible := existencia - prestados
	if req.CantidadPrestada > disponible {
		return nil, fmt.Errorf("stock insuficiente: solicitado %.2f, disponible %.2f", req.CantidadPrestada, disponible)
	}

	// Crear el préstamo
	prestamo := &domain.Prestamo{
		Referencia:               req.Referencia,
		Descripcion:              req.Descripcion,
		Ext1:                     req.Ext1,
		UM:                       req.UM,
		CedulaOperario:           req.CedulaOperario,
		NombreOperario:           req.NombreOperario,
		CantidadPrestada:         req.CantidadPrestada,
		Estado:                   "activo",
		IDUsuarioPrestamista:     idUsuario,
		NombreUsuarioPrestamista: nombreUsuario,
	}

	id, err := s.repoPrestamo.CrearPrestamo(ctx, prestamo)
	if err != nil {
		return nil, fmt.Errorf("error al crear préstamo: %w", err)
	}

	prestamo.IDPrestamo = id

	return &domain.RespuestaPrestamo{
		Prestamo: *prestamo,
		Mensaje:  "préstamo creado exitosamente",
	}, nil
}

// ObtenerPrestamo obtiene un préstamo por ID
func (s *servicioPrestamoCRUDImpl) ObtenerPrestamo(ctx context.Context, id int64) (*domain.Prestamo, error) {
	return s.repoPrestamo.ObtenerPrestamoPorID(ctx, id)
}

// ListarPrestamos lista préstamos con filtros
func (s *servicioPrestamoCRUDImpl) ListarPrestamos(ctx context.Context, estado, cedula, referencia string) (*domain.ListadoPrestamosResponse, error) {
	prestamos, err := s.repoPrestamo.ListarPrestamos(ctx, estado, cedula, referencia)
	if err != nil {
		return nil, err
	}
	return &domain.ListadoPrestamosResponse{
		Total:     len(prestamos),
		Prestamos: prestamos,
	}, nil
}

// ListarPrestamosPorOperario lista préstamos de un operario
func (s *servicioPrestamoCRUDImpl) ListarPrestamosPorOperario(ctx context.Context, cedula string, estado string) (*domain.ListadoPrestamosResponse, error) {
	prestamos, err := s.repoPrestamo.ListarPrestamosPorOperario(ctx, cedula, estado)
	if err != nil {
		return nil, err
	}
	return &domain.ListadoPrestamosResponse{
		Total:     len(prestamos),
		Prestamos: prestamos,
	}, nil
}

// DevolverPrestamo registra la devolución de un préstamo
func (s *servicioPrestamoCRUDImpl) DevolverPrestamo(ctx context.Context, id int64, req domain.DevolverPrestamoRequest, idUsuario int32) error {
	prestamo, err := s.repoPrestamo.ObtenerPrestamoPorID(ctx, id)
	if err != nil {
		return err
	}
	if prestamo == nil {
		return fmt.Errorf("préstamo no encontrado")
	}

	// Si ya está devuelto, no hacer nada
	if prestamo.Estado == "devuelto" {
		return fmt.Errorf("el préstamo ya está completamente devuelto")
	}

	// Si hay cantidad específica, es devolución parcial
	if req.CantidadDevuelta > 0 {
		devolucion := &domain.DevolucionParcial{
			IDPrestamo:        id,
			CantidadDevuelta:  req.CantidadDevuelta,
			Observaciones:     req.Observaciones,
			IDUsuarioRegistro: idUsuario,
		}

		if err := s.repoPrestamo.CrearDevolucionParcial(ctx, devolucion); err != nil {
			return fmt.Errorf("error al registrar devolución parcial: %w", err)
		}

		// Calcular total devuelto
		totalDevuelto, err := s.repoPrestamo.SumarDevolucionesPorPrestamo(ctx, id)
		if err != nil {
			return err
		}

		// Determinar nuevo estado
		nuevoEstado := "parcial"
		var fechaDevolucion *time.Time = nil
		if totalDevuelto >= prestamo.CantidadPrestada {
			nuevoEstado = "devuelto"
			ahora := utils.AhoraBogota()
			fechaDevolucion = &ahora
		}

		return s.repoPrestamo.ActualizarEstadoPrestamo(ctx, id, nuevoEstado, fechaDevolucion, req.CondicionDevolucion, req.Observaciones)
	}

	// Devolución completa
	ahora := utils.AhoraBogota()
	return s.repoPrestamo.ActualizarEstadoPrestamo(ctx, id, "devuelto", &ahora, req.CondicionDevolucion, req.Observaciones)
}

// CalcularStockDisponible calcula el stock disponible de una referencia
func (s *servicioPrestamoCRUDImpl) CalcularStockDisponible(ctx context.Context, referencia string, existenciaUNOEE float64) (*domain.StockDisponibleResponse, error) {
	prestados, err := s.repoPrestamo.SumarCantidadPrestadaPorReferencia(ctx, referencia)
	if err != nil {
		return nil, err
	}

	return &domain.StockDisponibleResponse{
		Referencia: referencia,
		Existencia: existenciaUNOEE,
		Prestados:  prestados,
		Disponible: existenciaUNOEE - prestados,
	}, nil
}
