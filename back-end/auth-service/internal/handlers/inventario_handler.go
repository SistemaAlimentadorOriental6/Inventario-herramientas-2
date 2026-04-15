package handlers

import (
	"auth-service/internal/domain"
	"auth-service/internal/service"
	"encoding/json"
	"net/http"
)

type ManejadorInventario struct {
	svcInventario service.ServicioInventario
}

func NuevoManejadorInventario(svc service.ServicioInventario) *ManejadorInventario {
	return &ManejadorInventario{svcInventario: svc}
}

// GuardarInventario godoc
// POST /inventario/guardar
func (h *ManejadorInventario) GuardarInventario(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		responderError(w, http.StatusMethodNotAllowed, "método no permitido (usar POST)")
		return
	}

	var registros []domain.RegistroInventario
	if err := json.NewDecoder(r.Body).Decode(&registros); err != nil {
		responderError(w, http.StatusBadRequest, "cuerpo de petición inválido (se esperaba un array de registros)")
		return
	}

	if len(registros) == 0 {
		responderError(w, http.StatusBadRequest, "no se enviaron registros para guardar")
		return
	}

	procesados, err := h.svcInventario.ProcesarGuardado(r.Context(), registros)
	if err != nil {
		responderError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(procesados)
}
