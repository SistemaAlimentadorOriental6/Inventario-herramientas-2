package repository

import (
	"auth-service/internal/domain"
	"context"
	"database/sql"
	"fmt"
	"strings"
)

// repositorioUbicacionSQLServer implementa RepositorioUbicacion usando SQL Server
type repositorioUbicacionSQLServer struct {
	db *sql.DB
}

// NuevoRepositorioUbicacion crea la instancia del repositorio SQL Server
func NuevoRepositorioUbicacion(db *sql.DB) RepositorioUbicacion {
	return &repositorioUbicacionSQLServer{db: db}
}

// ObtenerFilasUbicaciones retorna todas las filas de vw_Ubicaciones con todos los campos necesarios
func (r *repositorioUbicacionSQLServer) ObtenerFilasUbicaciones(ctx context.Context) ([]domain.FilaUbicacion, error) {
	query := `
		SELECT 
			ISNULL(Referencia, ''), 
			ISNULL(Ext1, ''), 
			ISNULL(Descripcion, ''), 
			ISNULL(UM, ''), 
			ISNULL(Existencia, 0), 
			ISNULL(DescUbicacion, ''), 
			ISNULL(Empleado, '') 
		FROM dbo.vw_Ubicaciones`

	filas, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error al consultar vw_Ubicaciones: %w", err)
	}
	defer filas.Close()

	var resultado []domain.FilaUbicacion
	for filas.Next() {
		var fila domain.FilaUbicacion
		err := filas.Scan(
			&fila.Referencia,
			&fila.Ext1,
			&fila.Descripcion,
			&fila.UM,
			&fila.Existencia,
			&fila.DescUbicacion,
			&fila.Empleado,
		)
		if err != nil {
			return nil, fmt.Errorf("error al leer fila de vw_Ubicaciones: %w", err)
		}
		resultado = append(resultado, fila)
	}

	return resultado, filas.Err()
}

// ObtenerUsuariosUnicos retorna los usuarios únicos de vw_Ubicaciones parseando CEDULA-NOMBRE
func (r *repositorioUbicacionSQLServer) ObtenerUsuariosUnicos(ctx context.Context) ([]domain.UsuarioUbicacion, error) {
	query := `
		SELECT DISTINCT ISNULL(Empleado, '') 
		FROM dbo.vw_Ubicaciones 
		WHERE Empleado IS NOT NULL AND Empleado != ''`

	filas, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error al consultar usuarios de vw_Ubicaciones: %w", err)
	}
	defer filas.Close()

	var resultado []domain.UsuarioUbicacion
	for filas.Next() {
		var empleado string
		err := filas.Scan(&empleado)
		if err != nil {
			return nil, fmt.Errorf("error al leer empleado: %w", err)
		}

		cedula, nombre := parsearEmpleado(empleado)
		if cedula != "" && nombre != "" {
			resultado = append(resultado, domain.UsuarioUbicacion{
				Cedula: cedula,
				Nombre: nombre,
			})
		}
	}

	return resultado, filas.Err()
}

// ObtenerPorCedula busca un empleado en vw_Ubicaciones por su número de cédula
func (r *repositorioUbicacionSQLServer) ObtenerPorCedula(ctx context.Context, cedula string) (*domain.UsuarioUbicacion, error) {
	query := `
		SELECT DISTINCT TOP 1 ISNULL(Empleado, '') 
		FROM dbo.vw_Ubicaciones 
		WHERE Empleado LIKE @p1 AND Empleado IS NOT NULL AND Empleado != ''`

	var empleado string
	err := r.db.QueryRowContext(ctx, query, sql.Named("p1", cedula+"-%")).Scan(&empleado)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrUsuarioNoEncontrado
		}
		return nil, fmt.Errorf("error al buscar empleado por cédula: %w", err)
	}

	cedulaEncontrada, nombre := parsearEmpleado(empleado)
	if cedulaEncontrada == "" || nombre == "" {
		return nil, domain.ErrUsuarioNoEncontrado
	}

	return &domain.UsuarioUbicacion{
		Cedula: cedulaEncontrada,
		Nombre: nombre,
	}, nil
}

// parsearEmpleado separa el formato "71214445- TORRES SANCHEZ WILMAR ALEXANDER"
// en cedula y nombre. También maneja formatos como "71214445-TORRES SANCHEZ WILMAR ALEXANDER"
func parsearEmpleado(empleado string) (cedula string, nombre string) {
	empleado = strings.TrimSpace(empleado)
	if empleado == "" {
		return "", ""
	}

	// Buscar el separador "-" (puede tener espacios alrededor)
	partes := strings.SplitN(empleado, "-", 2)
	if len(partes) != 2 {
		return "", ""
	}

	cedula = strings.TrimSpace(partes[0])
	nombre = strings.TrimSpace(partes[1])

	// Validar que la cédula sea numérica
	if _, err := fmt.Sscanf(cedula, "%d", new(int)); err != nil {
		return "", ""
	}

	return cedula, nombre
}
