package service

import (
	"auth-service/internal/domain"
	"auth-service/internal/repository"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// ServicioAuth define el contrato de autenticación
type ServicioAuth interface {
	Login(ctx context.Context, peticion domain.PeticionLogin) (*domain.RespuestaLogin, error)
	LoginPorCedula(ctx context.Context, peticion domain.PeticionLoginCedula) (*domain.RespuestaLogin, error)
	ValidarToken(tokenStr string) (*domain.ReclaimsJWT, error)
	CrearUsuario(ctx context.Context, peticion domain.PeticionCrearUsuario) (*domain.RespuestaCrearUsuario, error)
}

// servicioAuthImpl implementa ServicioAuth
type servicioAuthImpl struct {
	repo               repository.RepositorioUsuario
	repoUbicacion      repository.RepositorioUbicacion
	jwtSecreto         []byte
	jwtExpiracionHoras int
}

// claimsJWT embebe los campos estándar de JWT + nuestros datos de usuario
type claimsJWT struct {
	IDUsuario int32             `json:"id_usuario"`
	Email     string            `json:"email"`
	Rol       domain.RolUsuario `json:"rol"`
	jwt.RegisteredClaims
}

// NuevoServicioAuth crea el servicio de autenticación con sus dependencias
func NuevoServicioAuth(repo repository.RepositorioUsuario, jwtSecreto string, expHoras int) ServicioAuth {
	return &servicioAuthImpl{
		repo:               repo,
		jwtSecreto:         []byte(jwtSecreto),
		jwtExpiracionHoras: expHoras,
	}
}

// NuevoServicioAuthConUbicacion crea el servicio con acceso a vw_Ubicaciones
func NuevoServicioAuthConUbicacion(repo repository.RepositorioUsuario, repoUbic repository.RepositorioUbicacion, jwtSecreto string, expHoras int) ServicioAuth {
	return &servicioAuthImpl{
		repo:               repo,
		repoUbicacion:      repoUbic,
		jwtSecreto:         []byte(jwtSecreto),
		jwtExpiracionHoras: expHoras,
	}
}

// Login valida las credenciales y retorna un JWT si son correctas
func (s *servicioAuthImpl) Login(ctx context.Context, peticion domain.PeticionLogin) (*domain.RespuestaLogin, error) {
	// 1. Buscar usuario por email
	usuario, contrasenaHash, err := s.repo.ObtenerPorEmail(ctx, peticion.Email)
	if err != nil {
		if errors.Is(err, domain.ErrUsuarioNoEncontrado) {
			return nil, domain.ErrCredencialesInvalidas
		}
		return nil, fmt.Errorf("error interno al buscar usuario: %w", err)
	}

	// 2. Verificar que el usuario esté activo
	if !usuario.Activo {
		return nil, domain.ErrUsuarioInactivo
	}

	// 3. Comparar contraseña en texto plano directamente
	if peticion.Contrasena != contrasenaHash {
		return nil, domain.ErrCredencialesInvalidas
	}

	// 4. Generar token JWT
	token, err := s.generarToken(usuario)
	if err != nil {
		return nil, fmt.Errorf("error al generar token: %w", err)
	}

	// 5. Determinar ruta de redirección según el rol
	ruta := resolverRutaPorRol(usuario.Rol)

	return &domain.RespuestaLogin{
		Token:          token,
		Usuario:        *usuario,
		RedireccionarA: ruta,
	}, nil
}

// resolverRutaPorRol retorna la ruta del dashboard correspondiente al rol del usuario
func resolverRutaPorRol(rol domain.RolUsuario) string {
	switch rol {
	case domain.RolAdmin:
		return "/dashboard/admin"
	default:
		return "/dashboard/stock"
	}
}

// ValidarToken verifica y decodifica un JWT, retornando sus claims
func (s *servicioAuthImpl) ValidarToken(tokenStr string) (*domain.ReclaimsJWT, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &claimsJWT{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("método de firma inesperado: %v", t.Header["alg"])
		}
		return s.jwtSecreto, nil
	})

	if err != nil || !token.Valid {
		return nil, domain.ErrTokenInvalido
	}

	claims, ok := token.Claims.(*claimsJWT)
	if !ok {
		return nil, domain.ErrTokenInvalido
	}

	return &domain.ReclaimsJWT{
		IDUsuario: claims.IDUsuario,
		Email:     claims.Email,
		Rol:       claims.Rol,
	}, nil
}

// generarToken crea un JWT firmado con los datos del usuario
func (s *servicioAuthImpl) generarToken(usuario *domain.Usuario) (string, error) {
	expiracion := time.Now().Add(time.Duration(s.jwtExpiracionHoras) * time.Hour)

	claims := &claimsJWT{
		IDUsuario: usuario.IDUsuario,
		Email:     usuario.Email,
		Rol:       usuario.Rol,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiracion),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   fmt.Sprintf("%d", usuario.IDUsuario),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecreto)
}

// LoginPorCedula valida la cédula contra vw_Ubicaciones y retorna JWT para visualizadores
func (s *servicioAuthImpl) LoginPorCedula(ctx context.Context, peticion domain.PeticionLoginCedula) (*domain.RespuestaLogin, error) {
	cedula := strings.TrimSpace(peticion.Cedula)
	if cedula == "" {
		return nil, fmt.Errorf("la cédula es requerida")
	}

	// Validar que la cédula sea numérica
	if _, err := fmt.Sscanf(cedula, "%d", new(int)); err != nil {
		return nil, fmt.Errorf("la cédula debe ser numérica")
	}

	// Verificar que el repositorio de ubicaciones esté disponible
	if s.repoUbicacion == nil {
		return nil, fmt.Errorf("servicio de ubicaciones no disponible")
	}

	// Buscar empleado por cédula
	usuarioUbic, err := s.repoUbicacion.ObtenerPorCedula(ctx, cedula)
	if err != nil {
		if errors.Is(err, domain.ErrUsuarioNoEncontrado) {
			return nil, domain.ErrCredencialesInvalidas
		}
		return nil, fmt.Errorf("error al buscar empleado: %w", err)
	}

	// Crear usuario virtual para el visualizador
	usuario := &domain.Usuario{
		IDUsuario:        int32(-1), // ID negativo para visualizadores
		Empleado:         usuarioUbic.Cedula,
		NombreCompleto:   usuarioUbic.Nombre,
		DescripcionCargo: "Empleado",
		Email:            cedula + "@visualizador.local",
		Rol:              domain.RolVisualizador,
		Activo:           true,
	}

	// Generar token JWT
	token, err := s.generarToken(usuario)
	if err != nil {
		return nil, fmt.Errorf("error al generar token: %w", err)
	}

	return &domain.RespuestaLogin{
		Token:          token,
		Usuario:        *usuario,
		RedireccionarA: "/dashboard/stock",
	}, nil
}

// CrearUsuario valida y registra un nuevo usuario en MySQL
func (s *servicioAuthImpl) CrearUsuario(ctx context.Context, peticion domain.PeticionCrearUsuario) (*domain.RespuestaCrearUsuario, error) {
	peticion.Empleado = strings.TrimSpace(peticion.Empleado)
	peticion.NombreCompleto = strings.TrimSpace(peticion.NombreCompleto)
	peticion.DescripcionCargo = strings.TrimSpace(peticion.DescripcionCargo)
	peticion.Email = strings.TrimSpace(strings.ToLower(peticion.Email))
	peticion.Contrasena = strings.TrimSpace(peticion.Contrasena)
	peticion.Rol = strings.TrimSpace(strings.ToLower(peticion.Rol))

	if peticion.Empleado == "" || peticion.NombreCompleto == "" || peticion.DescripcionCargo == "" || peticion.Email == "" || peticion.Contrasena == "" || peticion.Rol == "" {
		return nil, fmt.Errorf("todos los campos son requeridos")
	}

	if peticion.Rol != string(domain.RolOperario) && peticion.Rol != string(domain.RolAdmin) {
		return nil, fmt.Errorf("rol inválido: usar operario o admin")
	}

	usuario, err := s.repo.CrearUsuario(ctx, peticion)
	if err != nil {
		return nil, err
	}

	return &domain.RespuestaCrearUsuario{
		Mensaje: "usuario creado correctamente",
		Usuario: *usuario,
	}, nil
}
