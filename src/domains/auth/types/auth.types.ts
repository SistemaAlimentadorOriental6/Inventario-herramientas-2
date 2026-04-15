export interface CredencialesLogin {
  email: string
  contrasena: string
}

export interface CredencialesLoginCedula {
  cedula: string
}

export interface UsuarioAuth {
  id_usuario: number
  empleado: string
  nombre_completo: string
  descripcion_cargo: string
  email: string
  rol: 'admin' | 'operario' | 'supervisor' | 'visualizador'
  activo: boolean
}

export interface RespuestaLogin {
  token: string
  usuario: UsuarioAuth
  redireccionarA: string
}

export interface EstadoAuth {
  cargando: boolean
  error: string | null
  token: string | null
}
