import { defineStore } from 'pinia'
import { ref } from 'vue'
import type { EstadoAuth, RespuestaLogin, UsuarioAuth } from '../types/auth.types'
import { authService } from '../services/authService'
import type { CredencialesLogin, CredencialesLoginCedula } from '../types/auth.types'

const CLAVE_TOKEN   = 'inv_token'
const CLAVE_USUARIO = 'inv_usuario'

export const useAuthStore = defineStore('auth', () => {
  // Hidratar desde localStorage al iniciar para sobrevivir recargas
  const token   = ref<EstadoAuth['token']>(localStorage.getItem(CLAVE_TOKEN))
  const usuario = ref<UsuarioAuth | null>(
    JSON.parse(localStorage.getItem(CLAVE_USUARIO) ?? 'null')
  )
  const cargando = ref(false)
  const error    = ref<EstadoAuth['error']>(null)

  // Retorna la ruta de redirección indicada por el backend según el rol
  async function login(credenciales: CredencialesLogin): Promise<string> {
    cargando.value = true
    error.value = null
    try {
      const respuesta = await authService.iniciarSesion(credenciales)
      token.value   = respuesta.token
      usuario.value = respuesta.usuario
      // Persistir para sobrevivir recargas
      localStorage.setItem(CLAVE_TOKEN, respuesta.token)
      localStorage.setItem(CLAVE_USUARIO, JSON.stringify(respuesta.usuario))
      return respuesta.redireccionarA
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Error desconocido'
      return '/login'
    } finally {
      cargando.value = false
    }
  }

  // Login por cédula para empleados de vw_Ubicaciones (visualizadores)
  async function loginPorCedula(credenciales: CredencialesLoginCedula): Promise<string> {
    cargando.value = true
    error.value = null
    try {
      const respuesta = await authService.loginPorCedula(credenciales)
      token.value   = respuesta.token
      usuario.value = respuesta.usuario
      // Persistir para sobrevivir recargas
      localStorage.setItem(CLAVE_TOKEN, respuesta.token)
      localStorage.setItem(CLAVE_USUARIO, JSON.stringify(respuesta.usuario))
      return respuesta.redireccionarA
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'Cédula no encontrada'
      return '/login'
    } finally {
      cargando.value = false
    }
  }

  function cerrarSesion() {
    token.value   = null
    usuario.value = null
    error.value   = null
    localStorage.removeItem(CLAVE_TOKEN)
    localStorage.removeItem(CLAVE_USUARIO)
  }

  return { token, usuario, cargando, error, login, loginPorCedula, cerrarSesion }
})

