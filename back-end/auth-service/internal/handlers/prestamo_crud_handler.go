package handlers

import (
	"auth-service/internal/domain"
	"auth-service/internal/service"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

// ManejadorPrestamoCRUD maneja las peticiones HTTP para préstamos
type ManejadorPrestamoCRUD struct {
	svcPrestamo service.ServicioPrestamoCRUD
}

// NuevoManejadorPrestamoCRUD crea el handler de préstamos
func NuevoManejadorPrestamoCRUD(svc service.ServicioPrestamoCRUD) *ManejadorPrestamoCRUD {
	return &ManejadorPrestamoCRUD{svcPrestamo: svc}
}

// CrearPrestamo maneja POST /prestamos
func (h *ManejadorPrestamoCRUD) CrearPrestamo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		responderError(w, http.StatusMethodNotAllowed, "método no permitido")
		return
	}

	var req domain.CrearPrestamoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		responderError(w, http.StatusBadRequest, "error al decodificar JSON: "+err.Error())
		return
	}

	// Validaciones básicas
	if req.Referencia == "" || req.CedulaOperario == "" || req.NombreOperario == "" {
		responderError(w, http.StatusBadRequest, "referencia, cedula_operario y nombre_operario son obligatorios")
		return
	}
	if req.CantidadPrestada <= 0 {
		req.CantidadPrestada = 1
	}

	// Extraer usuario del contexto (seteado por middleware JWT)
	idUsuario := int32(0)
	nombreUsuario := ""
	if claims, ok := r.Context().Value("claims").(map[string]interface{}); ok {
		if id, ok := claims["id_usuario"].(float64); ok {
			idUsuario = int32(id)
		}
		if nombre, ok := claims["nombre"].(string); ok {
			nombreUsuario = nombre
		}
	}

	respuesta, err := h.svcPrestamo.CrearPrestamo(r.Context(), req, idUsuario, nombreUsuario)
	if err != nil {
		responderError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(respuesta)
}

// ListarPrestamos maneja GET /prestamos
func (h *ManejadorPrestamoCRUD) ListarPrestamos(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		responderError(w, http.StatusMethodNotAllowed, "método no permitido")
		return
	}

	// Query params
	estado := r.URL.Query().Get("estado")
	cedula := r.URL.Query().Get("cedula")
	referencia := r.URL.Query().Get("referencia")

	respuesta, err := h.svcPrestamo.ListarPrestamos(r.Context(), estado, cedula, referencia)
	if err != nil {
		responderError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(respuesta)
}

// ObtenerPrestamo maneja GET /prestamos/:id
func (h *ManejadorPrestamoCRUD) ObtenerPrestamo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		responderError(w, http.StatusMethodNotAllowed, "método no permitido")
		return
	}

	segmentos := obtenerSegmentosRuta(r.URL.Path)
	if len(segmentos) < 2 || segmentos[len(segmentos)-2] != "prestamos" {
		responderError(w, http.StatusBadRequest, "URL inválida")
		return
	}

	// Extraer ID de la URL (soporta /prestamos/:id y /api/prestamos/:id)
	idStr := segmentos[len(segmentos)-1]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		responderError(w, http.StatusBadRequest, "ID inválido")
		return
	}

	prestamo, err := h.svcPrestamo.ObtenerPrestamo(r.Context(), id)
	if err != nil {
		responderError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if prestamo == nil {
		responderError(w, http.StatusNotFound, "préstamo no encontrado")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(prestamo)
}

// ListarPrestamosPorOperario maneja GET /prestamos/operario/:cedula
func (h *ManejadorPrestamoCRUD) ListarPrestamosPorOperario(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		responderError(w, http.StatusMethodNotAllowed, "método no permitido")
		return
	}

	// Extraer cédula de la URL: /prestamos/operario/:cedula (con o sin prefijo /api)
	segmentos := obtenerSegmentosRuta(r.URL.Path)
	if len(segmentos) < 3 || segmentos[len(segmentos)-2] != "operario" {
		responderError(w, http.StatusBadRequest, "URL inválida")
		return
	}
	cedula := segmentos[len(segmentos)-1]

	estado := r.URL.Query().Get("estado")

	respuesta, err := h.svcPrestamo.ListarPrestamosPorOperario(r.Context(), cedula, estado)
	if err != nil {
		responderError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(respuesta)
}

// DevolverPrestamo maneja POST /prestamos/:id/devolver
func (h *ManejadorPrestamoCRUD) DevolverPrestamo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		responderError(w, http.StatusMethodNotAllowed, "método no permitido")
		return
	}

	// Extraer ID de la URL: /prestamos/:id/devolver (con o sin prefijo /api)
	segmentos := obtenerSegmentosRuta(r.URL.Path)
	if len(segmentos) < 3 || segmentos[len(segmentos)-1] != "devolver" {
		responderError(w, http.StatusBadRequest, "URL inválida")
		return
	}
	idStr := segmentos[len(segmentos)-2]
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		responderError(w, http.StatusBadRequest, "ID inválido")
		return
	}

	var req domain.DevolverPrestamoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Si no hay body, es devolución completa
		req = domain.DevolverPrestamoRequest{}
	}

	// Extraer usuario del contexto
	idUsuario := int32(0)
	if claims, ok := r.Context().Value("claims").(map[string]interface{}); ok {
		if id, ok := claims["id_usuario"].(float64); ok {
			idUsuario = int32(id)
		}
	}

	if err := h.svcPrestamo.DevolverPrestamo(r.Context(), id, req, idUsuario); err != nil {
		responderError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"mensaje": "devolución registrada exitosamente",
	})
}

func obtenerSegmentosRuta(path string) []string {
	path = strings.Trim(path, "/")
	if path == "" {
		return []string{}
	}

	return strings.Split(path, "/")
}
