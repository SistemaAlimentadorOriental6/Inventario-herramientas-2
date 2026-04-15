import type { 
  RespuestaCarritosAsignados, 
  RespuestaDetalladoCarrito, 
  RegistroInventario 
} from '../types/stock.types'

const API_URL = import.meta.env.VITE_API_URL ?? 'http://localhost:8080'

export const stockService = {
  async obtenerCarritosAsignados(idUsuario: number, token: string, cedula?: string): Promise<RespuestaCarritosAsignados> {
    // Para visualizadores se envía cédula, para admin/operario se envía id_usuario
    const body = cedula ? { cedula } : { id_usuario: idUsuario }

    const respuesta = await fetch(`${API_URL}/carritos/asignados`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      },
      body: JSON.stringify(body),
    })

    if (!respuesta.ok) {
      const error = await respuesta.json().catch(() => ({}))
      throw new Error(error.error ?? 'Error al obtener carritos')
    }

    return respuesta.json()
  },

  async obtenerDetalladoCarrito(idUsuario: number, numCarrito: number, token: string, cedula?: string): Promise<RespuestaDetalladoCarrito> {
    // Para visualizadores se envía cédula, para admin/operario se envía id_usuario
    const body = cedula
      ? { cedula, numero_carrito: numCarrito }
      : { id_usuario: idUsuario, numero_carrito: numCarrito }

    const respuesta = await fetch(`${API_URL}/carritos/detallado`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      },
      body: JSON.stringify(body),
    })

    if (!respuesta.ok) {
      const error = await respuesta.json().catch(() => ({}))
      throw new Error(error.error ?? 'Error al obtener detalle del carrito')
    }

    return respuesta.json()
  },

  async guardarInventario(registros: RegistroInventario[], token: string) {
    const respuesta = await fetch(`${API_URL}/inventario/guardar`, {
      method: 'POST',
      headers: { 
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      },
      body: JSON.stringify(registros),
    })

    if (!respuesta.ok) {
      const error = await respuesta.json().catch(() => ({}))
      throw new Error(error.error ?? 'Error al guardar inventario')
    }

    return respuesta.json()
  },

  async obtenerListadoPartes(token: string): Promise<any> {
    const respuesta = await fetch(`${API_URL}/carritos/listado-partes`, {
      method: 'GET',
      headers: { 
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      }
    })

    if (!respuesta.ok) {
      const error = await respuesta.json().catch(() => ({}))
      throw new Error(error.error ?? 'Error al obtener listado de partes')
    }

    return respuesta.json()
  }
}
