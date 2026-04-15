package repository

import (
	"auth-service/internal/db/sqlc"
	"auth-service/internal/domain"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

// repositorioUsuarioMySQL implementa RepositorioUsuario usando SQLC + MySQL
type repositorioUsuarioMySQL struct {
	db      *sql.DB
	queries sqlc.Querier
}

const queryCrearUsuario = `
	INSERT INTO usuarios (
		empleado,
		nombre_completo,
		descripcion_cargo,
		email,
		contrasena,
		rol,
		activo
	) VALUES (?, ?, ?, ?, ?, ?, 1)`

// NuevoRepositorioUsuario crea una instancia del repositorio MySQL
func NuevoRepositorioUsuario(db *sql.DB, queries sqlc.Querier) RepositorioUsuario {
	return &repositorioUsuarioMySQL{db: db, queries: queries}
}

// ObtenerPorEmail retorna el usuario y su contraseña hash para validación
func (r *repositorioUsuarioMySQL) ObtenerPorEmail(ctx context.Context, email string) (*domain.Usuario, string, error) {
	usuarioSQL, err := r.queries.ObtenerUsuarioPorEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, "", domain.ErrUsuarioNoEncontrado
		}
		return nil, "", fmt.Errorf("error al buscar usuario por email: %w", err)
	}

	usuario := mapearADominio(usuarioSQL)
	contrasenaHash := usuarioSQL.Contrasena.String
	return usuario, contrasenaHash, nil
}

// ObtenerPorId retorna el usuario por su ID sin exponer la contraseña
func (r *repositorioUsuarioMySQL) ObtenerPorId(ctx context.Context, id int32) (*domain.Usuario, error) {
	usuarioSQL, err := r.queries.ObtenerUsuarioPorId(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUsuarioNoEncontrado
		}
		return nil, fmt.Errorf("error al buscar usuario por id: %w", err)
	}

	return mapearADominio(usuarioSQL), nil
}

// CrearUsuario inserta un nuevo usuario activo en MySQL
func (r *repositorioUsuarioMySQL) CrearUsuario(ctx context.Context, peticion domain.PeticionCrearUsuario) (*domain.Usuario, error) {
	resultado, err := r.db.ExecContext(ctx, queryCrearUsuario,
		peticion.Empleado,
		peticion.NombreCompleto,
		peticion.DescripcionCargo,
		peticion.Email,
		peticion.Contrasena,
		peticion.Rol,
	)
	if err != nil {
		mensaje := strings.ToLower(err.Error())
		if strings.Contains(mensaje, "for key 'email'") || strings.Contains(mensaje, "for key 'empleado'") {
			return nil, domain.ErrUsuarioExistente
		}
		return nil, fmt.Errorf("error al crear usuario: %w", err)
	}

	idCreado, err := resultado.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("error obteniendo id del usuario creado: %w", err)
	}

	return r.ObtenerPorId(ctx, int32(idCreado))
}

// mapearADominio convierte el modelo SQLC al dominio de la aplicación
func mapearADominio(u sqlc.Usuario) *domain.Usuario {
	return &domain.Usuario{
		IDUsuario:          u.IDUsuario,
		Empleado:           u.Empleado.String,
		NombreCompleto:     u.NombreCompleto.String,
		DescripcionCargo:   u.DescripcionCargo.String,
		Email:              u.Email.String,
		Rol:                domain.RolUsuario(u.Rol),
		Activo:             u.Activo == 1,
		FechaCreacion:      u.FechaCreacion,
		FechaActualizacion: u.FechaActualizacion,
	}
}
