package repository

import (
	"context"
)

// RepositorioAdmon define el acceso a la base de datos de administración (MySQL ADMON)
type RepositorioAdmon interface {
	// ObtenerMarcas retorna el listado de marcas únicas de la tabla lista_partes
	ObtenerMarcas(ctx context.Context) ([]string, error)
	// ContarTotalPartes retorna el total de registros en la tabla lista_partes
	ContarTotalPartes(ctx context.Context) (int, error)
}
