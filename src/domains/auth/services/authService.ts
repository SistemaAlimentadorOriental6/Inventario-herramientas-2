import type { CredencialesLogin, CredencialesLoginCedula, RespuestaLogin } from '../types/auth.types'

const API_URL = import.meta.env.VITE_API_URL ?? 'http://localhost:8080'

export const authService = {
  async iniciarSesion(credenciales: CredencialesLogin): Promise<RespuestaLogin> {
    const respuesta = await fetch(`${API_URL}/auth/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(credenciales),
    })

    if (!respuesta.ok) {
      const error = await respuesta.json().catch(() => ({}))
      throw new Error(error.error ?? 'Error al iniciar sesión')
    }

    return respuesta.json()
  },

  async loginPorCedula(credenciales: CredencialesLoginCedula): Promise<RespuestaLogin> {
    const respuesta = await fetch(`${API_URL}/auth/login-cedula`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(credenciales),
    })

    if (!respuesta.ok) {
      const error = await respuesta.json().catch(() => ({}))
      throw new Error(error.error ?? 'Cédula no encontrada')
    }

    return respuesta.json()
  },
}
