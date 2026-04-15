package handlers

import (
	"auth-service/internal/service"
	"encoding/json"
	"net/http"
)

// ManejadorPrestamo gestiona los endpoints de préstamo
type ManejadorPrestamo struct {
	svcPrestamo service.ServicioPrestamo
}

// NuevoManejadorPrestamo crea el handler con su dependencia inyectada
func NuevoManejadorPrestamo(svc service.ServicioPrestamo) *ManejadorPrestamo {
	return &ManejadorPrestamo{svcPrestamo: svc}
}

// ItemsPrestamo retorna los ítems disponibles para préstamo desde UNOEE
// GET /prestamo/items
func (h *ManejadorPrestamo) ItemsPrestamo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		responderError(w, http.StatusMethodNotAllowed, "método no permitido (usar GET)")
		return
	}

	respuesta, err := h.svcPrestamo.ObtenerItemsPrestamo(r.Context())
	if err != nil {
		responderError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(respuesta)
}
