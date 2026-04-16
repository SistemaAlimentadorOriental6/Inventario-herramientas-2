package repository

import (
	"auth-service/internal/domain"
	"context"
	"database/sql"
	"fmt"
	"strings"
)

// repositorioUNOEESQLServer implementa RepositorioUNOEE usando SQL Server
type repositorioUNOEESQLServer struct {
	db *sql.DB
}

// NuevoRepositorioUNOEE crea la instancia del repositorio SQL Server
func NuevoRepositorioUNOEE(db *sql.DB) RepositorioUNOEE {
	return &repositorioUNOEESQLServer{db: db}
}

// ObtenerItemsPrestamo retorna los ítems de la vista ADMIN-INVENTARIO_TODAS_BODEGAS
// filtrados por bodega "BODEGA PRESTAMO HERR" y existencia > 0
func (r *repositorioUNOEESQLServer) ObtenerItemsPrestamo(ctx context.Context) ([]domain.ItemPrestamo, error) {
	query := `
		SELECT 
			LTRIM(RTRIM(f_referencia)), 
			ISNULL(f121_id_ext1_detalle, ''), 
			ISNULL(f_desc_item, ''), 
			ISNULL(f_um, ''), 
			ISNULL(f_cant_existencia_1, 0), 
			ISNULL(v400_bodega, '') 
		FROM dbo.[ADMIN-INVENTARIO_TODAS_BODEGAS] 
		WHERE v400_bodega = 'BODEGA PRESTAMO HERR' 
			AND f_cant_existencia_1 > 0`

	filas, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error al consultar ADMIN-INVENTARIO_TODAS_BODEGAS: %w", err)
	}
	defer filas.Close()

	var resultado []domain.ItemPrestamo
	for filas.Next() {
		var item domain.ItemPrestamo
		err := filas.Scan(
			&item.Referencia,
			&item.Marca,
			&item.Descripcion,
			&item.UM,
			&item.Existencia,
			&item.Bodega,
		)
		if err != nil {
			return nil, fmt.Errorf("error al leer fila de ADMIN-INVENTARIO_TODAS_BODEGAS: %w", err)
		}
		resultado = append(resultado, item)
	}

	return resultado, filas.Err()
}

func (r *repositorioUNOEESQLServer) ObtenerExistenciaPorReferencia(ctx context.Context, referencia string) (float64, error) {
	// Limpiar referencia de espacios
	refLimpia := strings.TrimSpace(referencia)

	query := `
		SELECT ISNULL(f_cant_existencia_1, 0)
		FROM dbo.[ADMIN-INVENTARIO_TODAS_BODEGAS]
		WHERE LTRIM(RTRIM(f_referencia)) = @p1 AND v400_bodega = 'BODEGA PRESTAMO HERR'`

	var existencia float64
	err := r.db.QueryRowContext(ctx, query, refLimpia).Scan(&existencia)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("referencia no encontrada en bodega de préstamo: [%s]", refLimpia)
		}
		return 0, fmt.Errorf("error al consultar existencia: %w", err)
	}
	return existencia, nil
}
