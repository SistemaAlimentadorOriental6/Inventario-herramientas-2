package repository

import (
	"auth-service/internal/domain"
	"context"
	"database/sql"
	"fmt"
	"time"
)

// repositorioPrestamoMySQL implementa RepositorioPrestamo usando MySQL
type repositorioPrestamoMySQL struct {
	db *sql.DB
}

// NuevoRepositorioPrestamo crea la instancia del repositorio MySQL
func NuevoRepositorioPrestamo(db *sql.DB) RepositorioPrestamo {
	return &repositorioPrestamoMySQL{db: db}
}

// CrearPrestamo crea un nuevo préstamo
func (r *repositorioPrestamoMySQL) CrearPrestamo(ctx context.Context, p *domain.Prestamo) (int64, error) {
	query := `
		INSERT INTO prestamos (
			referencia, descripcion, ext1, um,
			cedula_operario, nombre_operario, cantidad_prestada,
			estado, id_usuario_prestamista, nombre_usuario_prestamista
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := r.db.ExecContext(ctx, query,
		p.Referencia, p.Descripcion, p.Ext1, p.UM,
		p.CedulaOperario, p.NombreOperario, p.CantidadPrestada,
		p.Estado, p.IDUsuarioPrestamista, p.NombreUsuarioPrestamista,
	)
	if err != nil {
		return 0, fmt.Errorf("error al crear préstamo: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("error al obtener ID del préstamo: %w", err)
	}

	return id, nil
}

// ObtenerPrestamoPorID obtiene un préstamo por su ID
func (r *repositorioPrestamoMySQL) ObtenerPrestamoPorID(ctx context.Context, id int64) (*domain.Prestamo, error) {
	query := `
		SELECT id_prestamo, referencia, descripcion, ext1, um,
			cedula_operario, nombre_operario, cantidad_prestada,
			estado, fecha_prestamo, fecha_devolucion,
			id_usuario_prestamista, nombre_usuario_prestamista,
			condicion_devolucion, observaciones_devolucion,
			created_at, updated_at
		FROM prestamos WHERE id_prestamo = ?`

	var p domain.Prestamo
	var fechaDevolucion sql.NullTime
	var condicion sql.NullString
	var observacion sql.NullString
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&p.IDPrestamo, &p.Referencia, &p.Descripcion, &p.Ext1, &p.UM,
		&p.CedulaOperario, &p.NombreOperario, &p.CantidadPrestada,
		&p.Estado, &p.FechaPrestamo, &fechaDevolucion,
		&p.IDUsuarioPrestamista, &p.NombreUsuarioPrestamista,
		&condicion, &observacion,
		&p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("error al obtener préstamo: %w", err)
	}

	if fechaDevolucion.Valid {
		p.FechaDevolucion = &fechaDevolucion.Time
	}
	if condicion.Valid {
		p.CondicionDevolucion = condicion.String
	}
	if observacion.Valid {
		p.ObservacionesDevolucion = observacion.String
	}

	return &p, nil
}

// ListarPrestamos lista préstamos con filtros opcionales
func (r *repositorioPrestamoMySQL) ListarPrestamos(ctx context.Context, estado, cedula, referencia string) ([]domain.Prestamo, error) {
	query := `
		SELECT id_prestamo, TRIM(referencia), descripcion, ext1, um,
			cedula_operario, nombre_operario, cantidad_prestada,
			estado, fecha_prestamo, fecha_devolucion,
			id_usuario_prestamista, nombre_usuario_prestamista,
			condicion_devolucion, observaciones_devolucion,
			created_at, updated_at
		FROM prestamos WHERE 1=1`
	var args []interface{}

	if estado != "" {
		query += " AND estado = ?"
		args = append(args, estado)
	}
	if cedula != "" {
		query += " AND cedula_operario = ?"
		args = append(args, cedula)
	}
	if referencia != "" {
		query += " AND referencia = ?"
		args = append(args, referencia)
	}
	query += " ORDER BY fecha_prestamo DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error al listar préstamos: %w", err)
	}
	defer rows.Close()

	return r.scanPrestamos(rows)
}

// ListarPrestamosPorOperario lista préstamos de un operario específico
func (r *repositorioPrestamoMySQL) ListarPrestamosPorOperario(ctx context.Context, cedula string, estado string) ([]domain.Prestamo, error) {
	query := `
		SELECT id_prestamo, TRIM(referencia), descripcion, ext1, um,
			cedula_operario, nombre_operario, cantidad_prestada,
			estado, fecha_prestamo, fecha_devolucion,
			id_usuario_prestamista, nombre_usuario_prestamista,
			condicion_devolucion, observaciones_devolucion,
			created_at, updated_at
		FROM prestamos WHERE cedula_operario = ?`
	args := []interface{}{cedula}

	if estado != "" {
		query += " AND estado = ?"
		args = append(args, estado)
	}
	query += " ORDER BY fecha_prestamo DESC"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error al listar préstamos del operario: %w", err)
	}
	defer rows.Close()

	return r.scanPrestamos(rows)
}

// scanPrestamos escanea las filas de préstamos
func (r *repositorioPrestamoMySQL) scanPrestamos(rows *sql.Rows) ([]domain.Prestamo, error) {
	var prestamos []domain.Prestamo
	for rows.Next() {
		var p domain.Prestamo
		var fechaDevolucion sql.NullTime
		var condicion sql.NullString
		var observacion sql.NullString
		err := rows.Scan(
			&p.IDPrestamo, &p.Referencia, &p.Descripcion, &p.Ext1, &p.UM,
			&p.CedulaOperario, &p.NombreOperario, &p.CantidadPrestada,
			&p.Estado, &p.FechaPrestamo, &fechaDevolucion,
			&p.IDUsuarioPrestamista, &p.NombreUsuarioPrestamista,
			&condicion, &observacion,
			&p.CreatedAt, &p.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error al escanear préstamo: %w", err)
		}
		if fechaDevolucion.Valid {
			p.FechaDevolucion = &fechaDevolucion.Time
		}
		if condicion.Valid {
			p.CondicionDevolucion = condicion.String
		}
		if observacion.Valid {
			p.ObservacionesDevolucion = observacion.String
		}
		prestamos = append(prestamos, p)
	}
	return prestamos, rows.Err()
}

// ActualizarEstadoPrestamo actualiza el estado y opcionalmente la fecha de devolución y observaciones
func (r *repositorioPrestamoMySQL) ActualizarEstadoPrestamo(ctx context.Context, id int64, estado string, fechaDevolucion *time.Time, condicion string, observacion string) error {
	query := "UPDATE prestamos SET estado = ?"
	args := []interface{}{estado}

	if fechaDevolucion != nil {
		query += ", fecha_devolucion = ?"
		args = append(args, *fechaDevolucion)
	}

	if condicion != "" {
		query += ", condicion_devolucion = ?"
		args = append(args, condicion)
	}

	if observacion != "" {
		query += ", observaciones_devolucion = ?"
		args = append(args, observacion)
	}

	query += " WHERE id_prestamo = ?"
	args = append(args, id)

	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("error al actualizar estado del préstamo: %w", err)
	}
	return nil
}

// SumarCantidadPrestadaPorReferencia suma las cantidades prestadas activas de una referencia
func (r *repositorioPrestamoMySQL) SumarCantidadPrestadaPorReferencia(ctx context.Context, referencia string) (float64, error) {
	query := `
		SELECT COALESCE(SUM(cantidad_prestada), 0)
		FROM prestamos
		WHERE referencia = ? AND estado IN ('activo', 'parcial')`

	var suma float64
	err := r.db.QueryRowContext(ctx, query, referencia).Scan(&suma)
	if err != nil {
		return 0, fmt.Errorf("error al sumar cantidad prestada: %w", err)
	}
	return suma, nil
}

// CrearDevolucionParcial registra una devolución parcial
func (r *repositorioPrestamoMySQL) CrearDevolucionParcial(ctx context.Context, d *domain.DevolucionParcial) error {
	query := `
		INSERT INTO prestamo_devoluciones (id_prestamo, cantidad_devuelta, observaciones, id_usuario_registro)
		VALUES (?, ?, ?, ?)`

	_, err := r.db.ExecContext(ctx, query, d.IDPrestamo, d.CantidadDevuelta, d.Observaciones, d.IDUsuarioRegistro)
	if err != nil {
		return fmt.Errorf("error al crear devolución parcial: %w", err)
	}
	return nil
}

// SumarDevolucionesPorPrestamo suma las cantidades devueltas de un préstamo
func (r *repositorioPrestamoMySQL) SumarDevolucionesPorPrestamo(ctx context.Context, idPrestamo int64) (float64, error) {
	query := `
		SELECT COALESCE(SUM(cantidad_devuelta), 0)
		FROM prestamo_devoluciones
		WHERE id_prestamo = ?`

	var suma float64
	err := r.db.QueryRowContext(ctx, query, idPrestamo).Scan(&suma)
	if err != nil {
		return 0, fmt.Errorf("error al sumar devoluciones: %w", err)
	}
	return suma, nil
}
