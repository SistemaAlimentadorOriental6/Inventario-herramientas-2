package handlers

import (
	"auth-service/internal/domain"
	"auth-service/internal/service"
	"encoding/json"
	"net/http"
)

// ManejadorCarritos gestiona los endpoints de carritos asignados
type ManejadorCarritos struct {
	svcCarrito service.ServicioCarrito
}

// NuevoManejadorCarritos crea el handler con su dependencia inyectada
func NuevoManejadorCarritos(svc service.ServicioCarrito) *ManejadorCarritos {
	return &ManejadorCarritos{svcCarrito: svc}
}

// CarritosAsignados retorna los carritos asignados en MySQL y el total de registros en SQL Server
// POST /carritos/asignados - Acepta id_usuario (admin/operario) o cedula (visualizador)
func (h *ManejadorCarritos) CarritosAsignados(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		responderError(w, http.StatusMethodNotAllowed, "método no permitido (usar POST)")
		return
	}

	var req struct {
		IDUsuario int32  `json:"id_usuario"`
		Cedula    string `json:"cedula"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		responderError(w, http.StatusBadRequest, "cuerpo de petición inválido (JSON esperado)")
		return
	}

	var respuesta *domain.RespuestaCarritosAsignados
	var err error

	// Si viene cédula, buscar por cédula (visualizador)
	if req.Cedula != "" {
		respuesta, err = h.svcCarrito.ObtenerCarritosPorCedula(r.Context(), req.Cedula)
	} else if req.IDUsuario > 0 {
		// Si viene id_usuario válido, buscar normalmente (admin/operario)
		respuesta, err = h.svcCarrito.ObtenerCarritosAsignados(r.Context(), req.IDUsuario)
	} else {
		responderError(w, http.StatusBadRequest, "se requiere id_usuario o cedula")
		return
	}

	if err != nil {
		responderError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(respuesta)
}

// AsignarCarrito asigna un carrito a un usuario destino y transfiere si ya estaba en otro usuario
// POST /carritos/asignar
func (h *ManejadorCarritos) AsignarCarrito(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		responderError(w, http.StatusMethodNotAllowed, "método no permitido (usar POST)")
		return
	}

	var req struct {
		IDUsuario     int32 `json:"id_usuario"`
		NumeroCarrito int32 `json:"numero_carrito"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		responderError(w, http.StatusBadRequest, "cuerpo de petición inválido (JSON esperado)")
		return
	}

	if req.IDUsuario <= 0 || req.NumeroCarrito <= 0 {
		responderError(w, http.StatusBadRequest, "se requiere id_usuario y numero_carrito válidos")
		return
	}

	respuesta, err := h.svcCarrito.AsignarCarritoAUsuario(r.Context(), req.IDUsuario, req.NumeroCarrito)
	if err != nil {
		responderError(w, http.StatusBadRequest, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(respuesta)
}

// QuitarCarrito remueve un carrito de un usuario específico
// POST /carritos/quitar
func (h *ManejadorCarritos) QuitarCarrito(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		responderError(w, http.StatusMethodNotAllowed, "método no permitido (usar POST)")
		return
	}

	var req struct {
		IDUsuario     int32 `json:"id_usuario"`
		NumeroCarrito int32 `json:"numero_carrito"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		responderError(w, http.StatusBadRequest, "cuerpo de petición inválido (JSON esperado)")
		return
	}

	if req.IDUsuario <= 0 || req.NumeroCarrito <= 0 {
		responderError(w, http.StatusBadRequest, "se requiere id_usuario y numero_carrito válidos")
		return
	}

	respuesta, err := h.svcCarrito.QuitarCarritoDeUsuario(r.Context(), req.IDUsuario, req.NumeroCarrito)
	if err != nil {
		responderError(w, http.StatusBadRequest, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(respuesta)
}

// DetalladoCarrito retorna los ítems detallados de un carrito específico
// POST /carritos/detallado - Acepta (id_usuario + numero_carrito) o (cedula + numero_carrito)
func (h *ManejadorCarritos) DetalladoCarrito(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		responderError(w, http.StatusMethodNotAllowed, "método no permitido (usar POST)")
		return
	}

	var req struct {
		IDUsuario     int32  `json:"id_usuario"`
		NumeroCarrito int32  `json:"numero_carrito"`
		Cedula        string `json:"cedula"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		responderError(w, http.StatusBadRequest, "cuerpo de petición inválido (JSON esperado)")
		return
	}

	if req.NumeroCarrito <= 0 {
		responderError(w, http.StatusBadRequest, "se requiere numero_carrito válido")
		return
	}

	var items *domain.RespuestaDetalladoCarrito
	var err error

	// Si viene cédula, usar método alternativo sin validación de MySQL
	if req.Cedula != "" {
		items, err = h.svcCarrito.ObtenerDetalladoCarritoPorCedula(r.Context(), req.Cedula, req.NumeroCarrito)
	} else if req.IDUsuario > 0 {
		// Si viene id_usuario válido, usar método normal con validación
		items, err = h.svcCarrito.ObtenerDetalladoCarrito(r.Context(), req.IDUsuario, req.NumeroCarrito)
	} else {
		responderError(w, http.StatusBadRequest, "se requiere id_usuario o cedula")
		return
	}

	if err != nil {
		responderError(w, http.StatusForbidden, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

// ListadoPartes retorna el listado de marcas únicas desde ADMON
// GET /carritos/listado-partes
func (h *ManejadorCarritos) ListadoPartes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		responderError(w, http.StatusMethodNotAllowed, "método no permitido (usar GET)")
		return
	}

	marcas, err := h.svcCarrito.ObtenerListadoPartes(r.Context())
	if err != nil {
		responderError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(marcas)
}

// UsuariosConCarritos retorna los usuarios activos con sus carritos asignados
// GET /carritos/usuarios
func (h *ManejadorCarritos) UsuariosConCarritos(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		responderError(w, http.StatusMethodNotAllowed, "método no permitido (usar GET)")
		return
	}

	usuarios, err := h.svcCarrito.ObtenerUsuariosConCarritos(r.Context())
	if err != nil {
		responderError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(usuarios)
}

// CarritosGenerales retorna todos los carritos existentes con datos de persona
// GET /carritos/general
func (h *ManejadorCarritos) CarritosGenerales(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		responderError(w, http.StatusMethodNotAllowed, "método no permitido (usar GET)")
		return
	}

	respuesta, err := h.svcCarrito.ObtenerCarritosGenerales(r.Context())
	if err != nil {
		responderError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(respuesta)
}

// UsuariosUbicacion retorna los usuarios únicos de vw_Ubicaciones con cedula y nombre separados
// GET /carritos/usuarios-ubicacion
func (h *ManejadorCarritos) UsuariosUbicacion(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		responderError(w, http.StatusMethodNotAllowed, "método no permitido (usar GET)")
		return
	}

	respuesta, err := h.svcCarrito.ObtenerUsuariosUbicacion(r.Context())
	if err != nil {
		responderError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(respuesta)
}

// CumplimientoUsuarios retorna el porcentaje de cumplimiento diario de todos los usuarios
// GET /carritos/cumplimiento
func (h *ManejadorCarritos) CumplimientoUsuarios(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		responderError(w, http.StatusMethodNotAllowed, "método no permitido (usar GET)")
		return
	}

	respuesta, err := h.svcCarrito.ObtenerCumplimientoUsuarios(r.Context())
	if err != nil {
		responderError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(respuesta)
}
