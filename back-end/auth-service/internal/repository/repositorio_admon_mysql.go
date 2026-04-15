package repository

import (
	"context"
	"database/sql"
	"fmt"
)

type repositorioAdmonMySQL struct {
	db *sql.DB
}

// NuevoRepositorioAdmon crea una nueva instancia del repositorio MySQL para ADMON
func NuevoRepositorioAdmon(db *sql.DB) RepositorioAdmon {
	return &repositorioAdmonMySQL{db: db}
}

func (r *repositorioAdmonMySQL) ObtenerMarcas(ctx context.Context) ([]string, error) {
	query := "SELECT DISTINCT marca FROM lista_partes WHERE marca IS NOT NULL ORDER BY marca ASC;"
	
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error al consultar marcas en ADMON: %w", err)
	}
	defer rows.Close()

	var marcas []string
	for rows.Next() {
		var marca string
		if err := rows.Scan(&marca); err != nil {
			return nil, fmt.Errorf("error al escanear marca: %w", err)
		}
		marcas = append(marcas, marca)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error recorriendo resultados de marcas: %w", err)
	}

	return marcas, nil
}

func (r *repositorioAdmonMySQL) ContarTotalPartes(ctx context.Context) (int, error) {
	query := "SELECT COUNT(DISTINCT marca) FROM lista_partes WHERE marca IS NOT NULL;"
	
	var total int
	err := r.db.QueryRowContext(ctx, query).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("error al contar partes en ADMON: %w", err)
	}
	
	return total, nil
}
