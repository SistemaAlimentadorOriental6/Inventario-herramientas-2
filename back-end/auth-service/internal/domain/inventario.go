package domain

import "time"

// RegistroInventario representa la información a guardar en la tabla registros_inventario
type RegistroInventario struct {
	IDUsuario           int32     `json:"id_usuario"`
	NumeroCarrito       int32     `json:"numero_carrito"`
	NombreCarrito       string    `json:"nombre_carrito"`
	IDProducto          string    `json:"id_producto"`
	Referencia          string    `json:"referencia_producto"`
	Descripcion         string    `json:"descripcion_producto"`
	MarcaAdicional      string    `json:"marca_adicional"`
	Marca               string    `json:"marca"`
	CantidadSistema     float64   `json:"cantidad_sistema"`
	CantidadFisica      float64   `json:"cantidad_fisica"`
	UnidadMedida        string    `json:"unidad_medida"`
	Novedad             string    `json:"novedad"`           // 'ninguna', 'desgaste', 'faltante'
	AccionFaltante      string    `json:"accion_faltante"`   // 'ninguna', 'descuento', 'compra'
	Observacion         string    `json:"observacion"`
	FirmaDigital        string    `json:"firma_digital"`     // Base64
	FechaRegistro       time.Time `json:"fecha_registro"`
}

// RespuestaGuardado devuelve el estado del proceso de guardado
type RespuestaGuardado struct {
	Mensaje    string               `json:"mensaje"`
	Procesados int                  `json:"procesados"`
	Registros  []RegistroInventario `json:"registros"`
}
