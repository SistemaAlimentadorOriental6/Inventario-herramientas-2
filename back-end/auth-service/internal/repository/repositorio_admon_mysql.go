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

func (r *repositorioAdmonMySQL) ObtenerNombresPorReferencia(ctx context.Context, referencias []string) (map[string]string, error) {
	if len(referencias) == 0 {
		return make(map[string]string), nil
	}

	// Construir query dinámica para IN
	query := "SELECT referencia_inteligente, nombre FROM lista_partes WHERE referencia_inteligente IN ("
	args := make([]interface{}, len(referencias))
	for i, ref := range referencias {
		query += "?"
		if i < len(referencias)-1 {
			query += ","
		}
		args[i] = ref
	}
	query += ")"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error al consultar nombres por referencia en ADMON: %w", err)
	}
	defer rows.Close()

	nombres := make(map[string]string)
	for rows.Next() {
		var ref, nombre string
		if err := rows.Scan(&ref, &nombre); err != nil {
			return nil, fmt.Errorf("error al escanear nombre de ADMON: %w", err)
		}
		nombres[ref] = nombre
	}

	return nombres, nil
}
