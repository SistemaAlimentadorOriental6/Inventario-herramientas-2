package repository

import (
	"auth-service/internal/domain"
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// repositorioCarritoMySQL implementa RepositorioCarrito usando MySQL
type repositorioCarritoMySQL struct {
	db *sql.DB
}

// NuevoRepositorioCarrito crea la instancia del repositorio MySQL de carritos
func NuevoRepositorioCarrito(db *sql.DB) RepositorioCarrito {
	return &repositorioCarritoMySQL{db: db}
}

const queryCarritosPorUsuario = `
	SELECT numero_carrito
	FROM asignaciones_carritos
	WHERE id_usuario = ?`

const queryUsuariosConCarritos = `
	SELECT
		u.id_usuario,
		u.nombre_completo,
		u.email,
		ac.numero_carrito
	FROM usuarios u
	LEFT JOIN asignaciones_carritos ac ON ac.id_usuario = u.id_usuario
	WHERE u.activo = 1
	  AND u.rol = 'operario'
	ORDER BY u.id_usuario, ac.numero_carrito`

const queryExisteUsuarioOperarioActivo = `
	SELECT 1
	FROM usuarios
	WHERE id_usuario = ?
	  AND activo = 1
	  AND rol = 'operario'
	LIMIT 1`

// ObtenerCarritosPorUsuario retorna los números de carrito asignados al usuario en MySQL
func (r *repositorioCarritoMySQL) ObtenerCarritosPorUsuario(ctx context.Context, idUsuario int32) ([]int32, error) {
	filas, err := r.db.QueryContext(ctx, queryCarritosPorUsuario, idUsuario)
	if err != nil {
		return nil, fmt.Errorf("error al consultar carritos del usuario %d: %w", idUsuario, err)
	}
	defer filas.Close()

	var carritos []int32
	for filas.Next() {
		var num int32
		if err := filas.Scan(&num); err != nil {
			return nil, fmt.Errorf("error al leer numero_carrito: %w", err)
		}
		carritos = append(carritos, num)
	}

	return carritos, filas.Err()
}

// ObtenerUsuariosConCarritos retorna usuarios activos con el listado de carritos asignados
func (r *repositorioCarritoMySQL) ObtenerUsuariosConCarritos(ctx context.Context) ([]domain.UsuarioConCarritos, error) {
	filas, err := r.db.QueryContext(ctx, queryUsuariosConCarritos)
	if err != nil {
		return nil, fmt.Errorf("error al consultar usuarios con carritos: %w", err)
	}
	defer filas.Close()

	resultado := make([]domain.UsuarioConCarritos, 0)
	indicePorUsuario := make(map[int32]int)

	for filas.Next() {
		var idUsuario int32
		var nombre sql.NullString
		var correo sql.NullString
		var numeroCarrito sql.NullInt64

		if err := filas.Scan(&idUsuario, &nombre, &correo, &numeroCarrito); err != nil {
			return nil, fmt.Errorf("error al leer usuarios con carritos: %w", err)
		}

		idx, existe := indicePorUsuario[idUsuario]
		if !existe {
			resultado = append(resultado, domain.UsuarioConCarritos{
				IDUsuario: idUsuario,
				Nombre:    nombre.String,
				Correo:    correo.String,
				Carritos:  []int32{},
			})
			idx = len(resultado) - 1
			indicePorUsuario[idUsuario] = idx
		}

		if numeroCarrito.Valid {
			resultado[idx].Carritos = append(resultado[idx].Carritos, int32(numeroCarrito.Int64))
		}
	}

	if err := filas.Err(); err != nil {
		return nil, fmt.Errorf("error iterando usuarios con carritos: %w", err)
	}

	return resultado, nil
}

// ExisteUsuarioOperarioActivo valida si el usuario existe, está activo y tiene rol operario
func (r *repositorioCarritoMySQL) ExisteUsuarioOperarioActivo(ctx context.Context, idUsuario int32) (bool, error) {
	var existe int
	err := r.db.QueryRowContext(ctx, queryExisteUsuarioOperarioActivo, idUsuario).Scan(&existe)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("error al validar usuario operario %d: %w", idUsuario, err)
	}

	return true, nil
}

// AsignarCarritoAUsuario asigna un carrito al usuario destino removiéndolo de cualquier usuario previo
func (r *repositorioCarritoMySQL) AsignarCarritoAUsuario(ctx context.Context, idUsuario int32, numeroCarrito int32) ([]int32, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("error al iniciar transacción de asignación: %w", err)
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	rows, err := tx.QueryContext(ctx, `
		SELECT DISTINCT id_usuario
		FROM asignaciones_carritos
		WHERE numero_carrito = ?
		FOR UPDATE`, numeroCarrito)
	if err != nil {
		return nil, fmt.Errorf("error al consultar dueños previos del carrito %d: %w", numeroCarrito, err)
	}

	usuariosPrevios := make([]int32, 0)
	for rows.Next() {
		var idPrevio int32
		if scanErr := rows.Scan(&idPrevio); scanErr != nil {
			rows.Close()
			return nil, fmt.Errorf("error al leer dueño previo del carrito %d: %w", numeroCarrito, scanErr)
		}
		usuariosPrevios = append(usuariosPrevios, idPrevio)
	}
	if closeErr := rows.Close(); closeErr != nil {
		return nil, fmt.Errorf("error al cerrar lectura de dueños previos: %w", closeErr)
	}

	if _, err = tx.ExecContext(ctx, `DELETE FROM asignaciones_carritos WHERE numero_carrito = ?`, numeroCarrito); err != nil {
		return nil, fmt.Errorf("error al limpiar asignaciones previas del carrito %d: %w", numeroCarrito, err)
	}

	if _, err = tx.ExecContext(ctx, `
		INSERT INTO asignaciones_carritos (id_usuario, numero_carrito)
		VALUES (?, ?)`, idUsuario, numeroCarrito); err != nil {
		return nil, fmt.Errorf("error al asignar carrito %d al usuario %d: %w", numeroCarrito, idUsuario, err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("error al confirmar asignación de carrito: %w", err)
	}

	return usuariosPrevios, nil
}

// QuitarCarritoDeUsuario elimina la relación de un carrito con un usuario específico
func (r *repositorioCarritoMySQL) QuitarCarritoDeUsuario(ctx context.Context, idUsuario int32, numeroCarrito int32) (bool, error) {
	res, err := r.db.ExecContext(ctx, `
		DELETE FROM asignaciones_carritos
		WHERE id_usuario = ? AND numero_carrito = ?`, idUsuario, numeroCarrito)
	if err != nil {
		return false, fmt.Errorf("error al quitar carrito %d del usuario %d: %w", numeroCarrito, idUsuario, err)
	}

	filas, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("error al obtener filas afectadas al quitar carrito: %w", err)
	}

	return filas > 0, nil
}
