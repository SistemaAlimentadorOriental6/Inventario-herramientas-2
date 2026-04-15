package handlers

import (
	"auth-service/internal/domain"
	"auth-service/internal/middleware"
	"auth-service/internal/service"
	"encoding/json"
	"errors"
	"net/http"
)

// ManejadorAuth gestiona los endpoints de autenticación
type ManejadorAuth struct {
	svcAuth service.ServicioAuth
}

// NuevoManejadorAuth crea el handler con su dependencia inyectada
func NuevoManejadorAuth(svcAuth service.ServicioAuth) *ManejadorAuth {
	return &ManejadorAuth{svcAuth: svcAuth}
}

// respuestaError es la estructura estándar de errores JSON
type respuestaError struct {
	Error string `json:"error"`
}

// LoginPorCedula godoc
// POST /auth/login-cedula
// Body: { "cedula": "1037618963" }
func (h *ManejadorAuth) LoginPorCedula(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		responderError(w, http.StatusMethodNotAllowed, "método no permitido")
		return
	}

	var peticion domain.PeticionLoginCedula
	if err := json.NewDecoder(r.Body).Decode(&peticion); err != nil {
		responderError(w, http.StatusBadRequest, "cuerpo de la petición inválido")
		return
	}

	if peticion.Cedula == "" {
		responderError(w, http.StatusBadRequest, "la cédula es requerida")
		return
	}

	respuesta, err := h.svcAuth.LoginPorCedula(r.Context(), peticion)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrCredencialesInvalidas):
			responderError(w, http.StatusUnauthorized, "cédula no encontrada")
		default:
			responderError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	responderJSON(w, http.StatusOK, respuesta)
}

// Login godoc
// POST /auth/login
// Body: { "email": "x@x.com", "contrasena": "abc123" }
func (h *ManejadorAuth) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		responderError(w, http.StatusMethodNotAllowed, "método no permitido")
		return
	}

	var peticion domain.PeticionLogin
	if err := json.NewDecoder(r.Body).Decode(&peticion); err != nil {
		responderError(w, http.StatusBadRequest, "cuerpo de la petición inválido")
		return
	}

	if peticion.Email == "" || peticion.Contrasena == "" {
		responderError(w, http.StatusBadRequest, "email y contraseña son requeridos")
		return
	}

	respuesta, err := h.svcAuth.Login(r.Context(), peticion)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrCredencialesInvalidas):
			responderError(w, http.StatusUnauthorized, err.Error())
		case errors.Is(err, domain.ErrUsuarioInactivo):
			responderError(w, http.StatusForbidden, err.Error())
		default:
			responderError(w, http.StatusInternalServerError, "error interno del servidor")
		}
		return
	}

	responderJSON(w, http.StatusOK, respuesta)
}

// Perfil godoc
// GET /auth/perfil  (requiere Bearer token)
func (h *ManejadorAuth) Perfil(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ObtenerClaimsDelContexto(r.Context())
	if !ok {
		responderError(w, http.StatusUnauthorized, "no autenticado")
		return
	}

	responderJSON(w, http.StatusOK, map[string]interface{}{
		"id_usuario": claims.IDUsuario,
		"email":      claims.Email,
		"rol":        claims.Rol,
	})
}

// VerificarToken godoc
// POST /auth/verificar  (verifica el token internamente, útil para otros microservicios)
func (h *ManejadorAuth) VerificarToken(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Token == "" {
		responderError(w, http.StatusBadRequest, "token requerido")
		return
	}

	claims, err := h.svcAuth.ValidarToken(body.Token)
	if err != nil {
		responderError(w, http.StatusUnauthorized, "token inválido o expirado")
		return
	}

	responderJSON(w, http.StatusOK, claims)
}

// CrearUsuario godoc
// POST /auth/usuarios (solo admin, requiere Bearer token)
func (h *ManejadorAuth) CrearUsuario(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		responderError(w, http.StatusMethodNotAllowed, "método no permitido")
		return
	}

	claims, ok := middleware.ObtenerClaimsDelContexto(r.Context())
	if !ok {
		responderError(w, http.StatusUnauthorized, "no autenticado")
		return
	}

	if claims.Rol != domain.RolAdmin {
		responderError(w, http.StatusForbidden, "solo administradores pueden crear usuarios")
		return
	}

	var peticion domain.PeticionCrearUsuario
	if err := json.NewDecoder(r.Body).Decode(&peticion); err != nil {
		responderError(w, http.StatusBadRequest, "cuerpo de la petición inválido")
		return
	}

	respuesta, err := h.svcAuth.CrearUsuario(r.Context(), peticion)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrUsuarioExistente):
			responderError(w, http.StatusConflict, err.Error())
		default:
			responderError(w, http.StatusBadRequest, err.Error())
		}
		return
	}

	responderJSON(w, http.StatusCreated, respuesta)
}

// Funciones auxiliares ------------------------------------

func responderJSON(w http.ResponseWriter, codigo int, datos interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(codigo)
	json.NewEncoder(w).Encode(datos)
}

func responderError(w http.ResponseWriter, codigo int, mensaje string) {
	responderJSON(w, codigo, respuestaError{Error: mensaje})
}
