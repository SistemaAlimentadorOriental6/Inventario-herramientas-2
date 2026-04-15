package domain

// DetalleCarrito muestra cuántos registros tiene un carrito específico en SQL Server
type DetalleCarrito struct {
	NumeroCarrito int32  `json:"numero_carrito"`
	Nombre        string `json:"nombre"`
	Registros     int    `json:"registros"`
	Completados   int    `json:"completados"`
}

// PeticionDetalleCarrito es el cuerpo esperado para el endpoint de detalle
type PeticionDetalleCarrito struct {
	IDUsuario     int32 `json:"id_usuario"`
	NumeroCarrito int32 `json:"numero_carrito"`
}

// RespuestaDetalladoCarrito envuelve la lista de ítems con el total de productos
type RespuestaDetalladoCarrito struct {
	TotalProductos int                    `json:"total_productos"`
	Items          []ItemCarritoDetallado `json:"items"`
}

// ItemCarritoDetallado representa un ítem dentro de un carrito (desde SQL Server)
type ItemCarritoDetallado struct {
	Referencia  string  `json:"referencia"`
	Ext1        string  `json:"ext1"`
	Descripcion string  `json:"descripcion"`
	UM          string  `json:"um"`
	Existencia  float64 `json:"existencia"`
	Completado  int     `json:"completado"` // 1 = sí fue guardado, 2 = no
}

// FilaUbicacion representa una fila de la vista vw_Ubicaciones en SQL Server con todos los campos necesarios
type FilaUbicacion struct {
	Referencia    string
	Ext1          string
	Descripcion   string
	UM            string
	Existencia    float64
	DescUbicacion string
	Empleado      string
}

// RespuestaCarritosAsignados es la respuesta del endpoint de carritos
type RespuestaCarritosAsignados struct {
	IDUsuario      int32            `json:"id_usuario"`
	Carritos       []DetalleCarrito `json:"carritos_asignados"`
	TotalRegistros int              `json:"total_registros_ubicacion"`
}

// RespuestaListadoPartes envuelve el listado de marcas y el total de herramientas
type RespuestaListadoPartes struct {
	TotalHerramientas int      `json:"total_herramientas"`
	Marcas            []string `json:"marcas"`
}

// UsuarioConCarritos representa un usuario con sus carritos asignados
type UsuarioConCarritos struct {
	IDUsuario int32   `json:"id_usuario"`
	Nombre    string  `json:"nombre"`
	Correo    string  `json:"correo"`
	Carritos  []int32 `json:"carritos"`
}

// CarritoGeneral representa un carrito existente con datos de persona
type CarritoGeneral struct {
	NumeroCarrito string `json:"numero_carrito"`
	Cedula        string `json:"cedula"`
	NombreCompleto string `json:"nombre_completo"`
}

// RespuestaCarritosGenerales envuelve el listado global de carritos
type RespuestaCarritosGenerales struct {
	TotalCarritos int              `json:"total_carritos"`
	Carritos      []CarritoGeneral `json:"carritos"`
}

// RespuestaAsignacionCarrito informa el resultado de asignar/transferir un carrito
type RespuestaAsignacionCarrito struct {
	IDUsuarioDestino int32   `json:"id_usuario_destino"`
	NumeroCarrito    int32   `json:"numero_carrito"`
	Transferido      bool    `json:"transferido"`
	UsuariosPrevios  []int32 `json:"usuarios_previos"`
	Mensaje          string  `json:"mensaje"`
}

// RespuestaQuitarCarrito resultado de quitar un carrito de un usuario
type RespuestaQuitarCarrito struct {
	IDUsuario     int32  `json:"id_usuario"`
	NumeroCarrito int32  `json:"numero_carrito"`
	Quitado       bool   `json:"quitado"`
	Mensaje       string `json:"mensaje"`
}

// ItemPrestamo representa un producto disponible para préstamo desde ADMIN-INVENTARIO_TODAS_BODEGAS
type ItemPrestamo struct {
	Referencia   string  `json:"f_referencia"`
	Marca        string  `json:"f121_id_ext1_detalle"`
	Descripcion  string  `json:"f_desc_item"`
	UM           string  `json:"f_um"`
	Existencia   float64 `json:"f_cant_existencia_1"`
	Bodega       string  `json:"v400_bodega"`
}

// RespuestaItemsPrestamo envuelve la lista de ítems disponibles para préstamo
type RespuestaItemsPrestamo struct {
	Total int            `json:"total"`
	Items []ItemPrestamo `json:"items"`
}

// UsuarioUbicacion representa un usuario extraído de la columna Empleado en vw_Ubicaciones
type UsuarioUbicacion struct {
	Cedula string `json:"cedula"`
	Nombre string `json:"nombre"`
}

// RespuestaUsuariosUbicacion envuelve la lista de usuarios únicos
type RespuestaUsuariosUbicacion struct {
	Total    int                `json:"total"`
	Usuarios []UsuarioUbicacion `json:"usuarios"`
}

// CumplimientoCarrito representa el estado de cumplimiento de un carrito específico
type CumplimientoCarrito struct {
	NumeroCarrito  int32   `json:"numero_carrito"`
	Registros      int     `json:"registros"`
	Completados    int     `json:"completados"`
	Pendientes     int     `json:"pendientes"`
	Porcentaje     float64 `json:"porcentaje"`
}

// CumplimientoUsuario agrupa el cumplimiento de todos los carritos de un usuario
type CumplimientoUsuario struct {
	IDUsuario        int32               `json:"id_usuario"`
	Nombre           string              `json:"nombre"`
	Correo           string              `json:"correo"`
	Carritos         []CumplimientoCarrito `json:"carritos"`
	TotalRegistros   int                 `json:"total_registros"`
	TotalCompletados int                 `json:"total_completados"`
	TotalPendientes  int                 `json:"total_pendientes"`
	PorcentajeGlobal float64             `json:"porcentaje_global"`
	Estado           string              `json:"estado"`
}

// RespuestaCumplimientoUsuarios envuelve la lista de cumplimiento
type RespuestaCumplimientoUsuarios struct {
	Total    int                   `json:"total"`
	Usuarios []CumplimientoUsuario `json:"usuarios"`
}
