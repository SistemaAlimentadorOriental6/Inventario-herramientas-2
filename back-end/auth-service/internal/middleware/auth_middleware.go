package middleware

import (
	"auth-service/internal/domain"
	"auth-service/internal/service"
	"context"
	"net/http"
	"strings"
)

// ClaveContextoUsuario es la clave para guardar datos en el contexto HTTP
type ClaveContexto string

const ClaveUsuarioContexto ClaveContexto = "usuario_jwt"

// AutenticarJWT valida el Bearer token en el header Authorization
func AutenticarJWT(svcAuth service.ServicioAuth) func(http.Handler) http.Handler {
	return func(siguiente http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			encabezado := r.Header.Get("Authorization")
			if encabezado == "" || !strings.HasPrefix(encabezado, "Bearer ") {
				http.Error(w, `{"error":"token requerido"}`, http.StatusUnauthorized)
				return
			}

			tokenStr := strings.TrimPrefix(encabezado, "Bearer ")
			claims, err := svcAuth.ValidarToken(tokenStr)
			if err != nil {
				http.Error(w, `{"error":"token inválido o expirado"}`, http.StatusUnauthorized)
				return
			}

			// Inyectamos los claims en el contexto para los handlers
			ctx := context.WithValue(r.Context(), ClaveUsuarioContexto, claims)
			siguiente.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// ObtenerClaimsDelContexto extrae los claims JWT del contexto HTTP
func ObtenerClaimsDelContexto(ctx context.Context) (*domain.ReclaimsJWT, bool) {
	claims, ok := ctx.Value(ClaveUsuarioContexto).(*domain.ReclaimsJWT)
	return claims, ok
}

// SoloRol permite el acceso únicamente a roles específicos
func SoloRol(roles ...domain.RolUsuario) func(http.Handler) http.Handler {
	return func(siguiente http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := ObtenerClaimsDelContexto(r.Context())
			if !ok {
				http.Error(w, `{"error":"no autenticado"}`, http.StatusUnauthorized)
				return
			}

			for _, rol := range roles {
				if claims.Rol == rol {
					siguiente.ServeHTTP(w, r)
					return
				}
			}

			http.Error(w, `{"error":"acceso denegado"}`, http.StatusForbidden)
		})
	}
}
