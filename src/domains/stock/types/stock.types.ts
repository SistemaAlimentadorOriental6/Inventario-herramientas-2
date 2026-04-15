export interface DetalleCarrito {
  numero_carrito: number;
  nombre: string;
  registros: number;
  completados: number;
}

export interface ItemCarritoDetallado {
  referencia: string;
  ext1: string;
  descripcion: string;
  um: string;
  existencia: number;
  completado: 1 | 2; // 1 = Sí, 2 = No
}

export interface RespuestaCarritosAsignados {
  id_usuario: number;
  carritos_asignados: DetalleCarrito[];
  total_registros_ubicacion: number;
}

export interface RespuestaDetalladoCarrito {
  total_productos: number;
  items: ItemCarritoDetallado[];
}

export interface RegistroInventario {
  id_usuario: number;
  numero_carrito: number;
  nombre_carrito: string;
  id_producto: string;
  referencia_producto: string;
  descripcion_producto: string;
  marca_adicional: string;
  marca: string;
  cantidad_sistema: number;
  cantidad_fisica: number;
  unidad_medida: string;
  novedad: 'ninguna' | 'desgaste' | 'faltante';
  accion_faltante: 'ninguna' | 'descuento' | 'compra';
  observacion: string;
  firma_digital: string; // Base64
}
