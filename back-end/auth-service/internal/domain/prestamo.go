package domain

import (
	"time"
)

// Prestamo representa un préstamo de herramienta a un operario
type Prestamo struct {
	IDPrestamo            int64     `json:"id_prestamo"`
	Referencia            string    `json:"referencia"`
	Descripcion           string    `json:"descripcion"`
	Ext1                  string    `json:"ext1"`
	UM                    string    `json:"um"`
	CedulaOperario        string    `json:"cedula_operario"`
	NombreOperario        string    `json:"nombre_operario"`
	CantidadPrestada      float64   `json:"cantidad_prestada"`
	Estado                string    `json:"estado"`
	FechaPrestamo         time.Time `json:"fecha_prestamo"`
	FechaDevolucion       *time.Time `json:"fecha_devolucion,omitempty"`
	IDUsuarioPrestamista  int32     `json:"id_usuario_prestamista"`
	NombreUsuarioPrestamista string `json:"nombre_usuario_prestamista"`
	CondicionDevolucion   string    `json:"condicion_devolucion,omitempty"`
	ObservacionesDevolucion string  `json:"observaciones_devolucion,omitempty"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

// DevolucionParcial representa una devolución parcial de un préstamo
type DevolucionParcial struct {
	IDDevolucion      int64     `json:"id_devolucion"`
	IDPrestamo        int64     `json:"id_prestamo"`
	CantidadDevuelta  float64   `json:"cantidad_devuelta"`
	FechaDevolucion   time.Time `json:"fecha_devolucion"`
	Observaciones     string    `json:"observaciones"`
	IDUsuarioRegistro int32     `json:"id_usuario_registro"`
}

// CrearPrestamoRequest datos para crear un préstamo
type CrearPrestamoRequest struct {
	Referencia       string  `json:"referencia"`
	Descripcion      string  `json:"descripcion"`
	Ext1             string  `json:"ext1"`
	UM               string  `json:"um"`
	CedulaOperario   string  `json:"cedula_operario"`
	NombreOperario   string  `json:"nombre_operario"`
	CantidadPrestada float64 `json:"cantidad_prestada"`
}

// DevolverPrestamoRequest datos para registrar devolución
type DevolverPrestamoRequest struct {
	CantidadDevuelta   float64 `json:"cantidad_devuelta,omitempty"`
	CondicionDevolucion string `json:"estado,omitempty"`
	Observaciones      string  `json:"observaciones,omitempty"`
}

// RespuestaPrestamo envuelve un préstamo creado
type RespuestaPrestamo struct {
	Prestamo Prestamo `json:"prestamo"`
	Mensaje  string   `json:"mensaje"`
}

// ListadoPrestamosResponse respuesta de listado
type ListadoPrestamosResponse struct {
	Total     int       `json:"total"`
	Prestamos []Prestamo `json:"prestamos"`
}

// StockDisponibleResponse para consultar stock
type StockDisponibleResponse struct {
	Referencia  string  `json:"referencia"`
	Existencia  float64 `json:"existencia"`
	Prestados   float64 `json:"prestados"`
	Disponible  float64 `json:"disponible"`
}
