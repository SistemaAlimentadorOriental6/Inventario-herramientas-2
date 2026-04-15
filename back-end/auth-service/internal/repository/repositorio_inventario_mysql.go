package repository

import (
	"auth-service/internal/domain"
	"context"
	"database/sql"
	"fmt"
	"strings"
)

type repositorioInventarioMySQL struct {
	db *sql.DB
}

func NuevoRepositorioInventario(db *sql.DB) RepositorioInventario {
	return &repositorioInventarioMySQL{db: db}
}

// GuardarRegistros inserta múltiples registros en una sola operación (Bulk Insert)
func (r *repositorioInventarioMySQL) GuardarRegistros(ctx context.Context, registros []domain.RegistroInventario) (int, error) {
	if len(registros) == 0 {
		return 0, nil
	}

	const cols = 15
	placeholder := "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"

	var sqlRaw strings.Builder
	sqlRaw.WriteString("INSERT INTO registros_inventario (id_usuario, numero_carrito, nombre_carrito, id_producto, referencia_producto, descripcion_producto, marca_adicional, marca, cantidad_sistema, cantidad_fisica, unidad_medida, novedad, accion_faltante, observacion, firma_digital) VALUES ")

	args := make([]interface{}, 0, len(registros)*cols)

	for i, reg := range registros {
		if i > 0 {
			sqlRaw.WriteString(",")
		}
		sqlRaw.WriteString(placeholder)
		args = append(args,
			reg.IDUsuario, reg.NumeroCarrito, reg.NombreCarrito, reg.IDProducto,
			strings.TrimSpace(reg.Referencia), reg.Descripcion, reg.MarcaAdicional,
			reg.Marca, reg.CantidadSistema, reg.CantidadFisica, reg.UnidadMedida,
			reg.Novedad, reg.AccionFaltante, reg.Observacion, reg.FirmaDigital,
		)
	}

	resultado, err := r.db.ExecContext(ctx, sqlRaw.String(), args...)
	if err != nil {
		return 0, fmt.Errorf("error al ejecutar inserción masiva: %w", err)
	}

	total, _ := resultado.RowsAffected()
	return int(total), nil
}

// ObtenerReferenciasGuardadas retorna un set de referencias ya guardadas para usuario+carrito
// Las referencias se guardan en TrimSpace para comparar correctamente con los datos de SQL Server
func (r *repositorioInventarioMySQL) ObtenerReferenciasGuardadas(ctx context.Context, idUsuario int32, numCarrito int32) (map[string]bool, error) {
	query := `SELECT TRIM(referencia_producto) FROM registros_inventario WHERE id_usuario = ? AND numero_carrito = ? AND DATE(fecha_registro) = CURDATE()`

	filas, err := r.db.QueryContext(ctx, query, idUsuario, numCarrito)
	if err != nil {
		return nil, fmt.Errorf("error al consultar referencias guardadas: %w", err)
	}
	defer filas.Close()

	set := make(map[string]bool)
	for filas.Next() {
		var ref string
		if err := filas.Scan(&ref); err != nil {
			return nil, err
		}
		set[strings.TrimSpace(ref)] = true
	}

	return set, filas.Err()
}

// ContarCompletadosPorUsuario retorna mapa de numero_carrito → cantidad de registros guardados
func (r *repositorioInventarioMySQL) ContarCompletadosPorUsuario(ctx context.Context, idUsuario int32) (map[int32]int, error) {
	query := `
		SELECT numero_carrito, COUNT(*) 
		FROM registros_inventario 
		WHERE id_usuario = ? AND DATE(fecha_registro) = CURDATE()
		GROUP BY numero_carrito`

	filas, err := r.db.QueryContext(ctx, query, idUsuario)
	if err != nil {
		return nil, fmt.Errorf("error al contar completados por carrito: %w", err)
	}
	defer filas.Close()

	mapa := make(map[int32]int)
	for filas.Next() {
		var numCarrito int32
		var total int
		if err := filas.Scan(&numCarrito, &total); err != nil {
			return nil, err
		}
		mapa[numCarrito] = total
	}

	return mapa, filas.Err()
}
