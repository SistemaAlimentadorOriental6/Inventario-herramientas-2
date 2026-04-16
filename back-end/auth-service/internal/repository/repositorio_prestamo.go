package repository

import (
	"auth-service/internal/domain"
	"context"
	"time"
)

// RepositorioPrestamo define las operaciones de acceso a datos para préstamos
type RepositorioPrestamo interface {
	// CrearPrestamo crea un nuevo préstamo
	CrearPrestamo(ctx context.Context, p *domain.Prestamo) (int64, error)

	// ObtenerPrestamoPorID obtiene un préstamo por su ID
	ObtenerPrestamoPorID(ctx context.Context, id int64) (*domain.Prestamo, error)

	// ListarPrestamos lista préstamos con filtros opcionales
	ListarPrestamos(ctx context.Context, estado, cedula, referencia string) ([]domain.Prestamo, error)

	// ListarPrestamosPorOperario lista préstamos de un operario específico
	ListarPrestamosPorOperario(ctx context.Context, cedula string, estado string) ([]domain.Prestamo, error)

	// ActualizarEstadoPrestamo actualiza el estado y opcionalmente la fecha de devolución y condiciones
	ActualizarEstadoPrestamo(ctx context.Context, id int64, estado string, fechaDevolucion *time.Time, condicion string, observacion string) error

	// SumarCantidadPrestadaPorReferencia suma las cantidades prestadas activas de una referencia
	SumarCantidadPrestadaPorReferencia(ctx context.Context, referencia string) (float64, error)

	// CrearDevolucionParcial registra una devolución parcial
	CrearDevolucionParcial(ctx context.Context, d *domain.DevolucionParcial) error

	// SumarDevolucionesPorPrestamo suma las cantidades devueltas de un préstamo
	SumarDevolucionesPorPrestamo(ctx context.Context, idPrestamo int64) (float64, error)
}
