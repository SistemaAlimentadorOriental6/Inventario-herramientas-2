<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../../auth/store/useAuthStore'

const API_URL = import.meta.env.VITE_API_URL ?? 'http://localhost:8080'

type UsuarioGestion = {
  id: number
  nombre: string
  email: string
  carritos: string[]
}

type UsuarioConCarritosApi = {
  id_usuario: number
  nombre: string
  correo: string
  carritos: Array<number | string>
}

type CumplimientoCarrito = {
  numero_carrito: number
  registros: number
  completados: number
  pendientes: number
  porcentaje: number
}

type CumplimientoUsuario = {
  id_usuario: number
  nombre: string
  correo: string
  carritos: CumplimientoCarrito[]
  total_registros: number
  total_completados: number
  total_pendientes: number
  porcentaje_global: number
  estado: 'pendiente' | 'sin_registros' | 'completado' | 'en_progreso'
}

type RespuestaCumplimientoApi = {
  total: number
  usuarios: CumplimientoUsuario[]
}

type CarritoGeneralApi = {
  numero_carrito: string
  cedula: string
  nombre_completo: string
}

type RespuestaCarritosGeneralesApi = {
  total_carritos: number
  carritos: CarritoGeneralApi[]
}

const router = useRouter()
const authStore = useAuthStore()

const usuarioLogueado = computed(() => authStore.usuario)
const esOperario = computed(() => usuarioLogueado.value?.rol === 'operario')

const personal = ref<UsuarioGestion[]>([])
const carritosCatalogo = ref<CarritoGeneralApi[]>([])
const cumplimientoUsuarios = ref<CumplimientoUsuario[]>([])
const cargandoDatos = ref(false)
const errorDatos = ref('')
const cargandoAccionModal = ref(false)
const carritoEnProceso = ref('')
const mensajeAccionModal = ref('')
const tipoMensajeAccionModal = ref<'ok' | 'error' | ''>('')
const cargandoCrearUsuario = ref(false)
const mensajeCrearUsuario = ref('')
const tipoMensajeCrearUsuario = ref<'ok' | 'error' | ''>('')

const busqueda = ref('')
const filtroEstado = ref('Todos')
const modalAbierto = ref(false)
const modalCrearAbierto = ref(false)
const personalSeleccionado = ref<UsuarioGestion | null>(null)
const tabActivo = ref('Carritos')
const busquedaCarritoModal = ref('')
const soloDisponibles = ref(false)

const nuevoUsuario = ref({
  empleado: '',
  nombre: '',
  cargo: '',
  email: '',
  clave: '',
  rol: 'operario'
})

const personalFiltrado = computed(() => {
  let filtrados = personal.value
  if (busqueda.value) {
    const q = busqueda.value.toLowerCase()
    filtrados = filtrados.filter(p => p.nombre.toLowerCase().includes(q) || p.email.toLowerCase().includes(q))
  }
  if (filtroEstado.value === 'Con carritos') {
    filtrados = filtrados.filter(p => p.carritos.length > 0)
  }
  if (filtroEstado.value === 'Sin carritos') {
    filtrados = filtrados.filter(p => p.carritos.length === 0)
  }
  return filtrados
})

const asignacionPorCarrito = computed(() => {
  const mapa = new Map<string, UsuarioGestion>()
  for (const persona of personal.value) {
    for (const carrito of persona.carritos) {
      mapa.set(String(carrito), persona)
    }
  }
  return mapa
})

const mapaCumplimiento = computed(() => {
  const mapa = new Map<number, CumplimientoUsuario>()
  for (const cumplimiento of cumplimientoUsuarios.value) {
    mapa.set(cumplimiento.id_usuario, cumplimiento)
  }
  return mapa
})

function obtenerCumplimientoUsuario(idUsuario: number): CumplimientoUsuario | undefined {
  return mapaCumplimiento.value.get(idUsuario)
}

const carritosUsuarioSeleccionado = computed(() => {
  if (!personalSeleccionado.value) return []
  return [...personalSeleccionado.value.carritos].sort((a, b) => a.localeCompare(b, undefined, { numeric: true }))
})

const carritosCatalogoFiltrados = computed(() => {
  const q = busquedaCarritoModal.value.trim().toLowerCase()
  return carritosCatalogo.value.filter((carrito) => {
    const textoBase = [carrito.numero_carrito, carrito.cedula, carrito.nombre_completo].join(' ').toLowerCase()
    const cumpleBusqueda = q === '' || textoBase.includes(q)
    if (!cumpleBusqueda) return false
    if (!soloDisponibles.value) return true
    return !asignacionPorCarrito.value.has(carrito.numero_carrito)
  })
})

function normalizarCarrito(valor: string | number): string {
  return String(valor).trim()
}

function convertirNumeroCarrito(numeroCarrito: string): number {
  const valor = Number(numeroCarrito)
  if (!Number.isFinite(valor) || valor <= 0) {
    throw new Error('El carrito a gestionar no tiene un número válido')
  }
  return valor
}

function obtenerTokenSeguro(): string {
  const token = authStore.token
  if (!token) {
    throw new Error('No hay sesión activa para consultar la gestión de carritos')
  }
  return token
}

async function cargarDatosGestion(opciones: { mostrarLoader?: boolean; mantenerSeleccionId?: number | null } = {}) {
  const mostrarLoader = opciones.mostrarLoader ?? true
  const mantenerSeleccionId = opciones.mantenerSeleccionId ?? personalSeleccionado.value?.id ?? null

  if (mostrarLoader) {
    cargandoDatos.value = true
  }
  errorDatos.value = ''
  try {
    const token = obtenerTokenSeguro()

    const respuestaUsuarios = await fetch(`${API_URL}/carritos/usuarios`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      }
    })

    if (!respuestaUsuarios.ok) {
      const error = await respuestaUsuarios.json().catch(() => ({}))
      throw new Error(error.error ?? 'Error al obtener usuarios')
    }

    const usuariosApi = await respuestaUsuarios.json() as UsuarioConCarritosApi[]
    personal.value = usuariosApi.map((u) => ({
      id: u.id_usuario,
      nombre: u.nombre,
      email: u.correo,
      carritos: (u.carritos ?? []).map(normalizarCarrito)
    }))

    const respuestaCarritos = await fetch(`${API_URL}/carritos/general`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      }
    })

    if (!respuestaCarritos.ok) {
      const error = await respuestaCarritos.json().catch(() => ({}))
      throw new Error(error.error ?? 'Error al obtener catálogo de carritos')
    }

    const dataCarritos = await respuestaCarritos.json() as RespuestaCarritosGeneralesApi | CarritoGeneralApi[]
    const carritos = Array.isArray(dataCarritos) ? dataCarritos : dataCarritos.carritos
    carritosCatalogo.value = (carritos ?? []).map((c) => ({
      numero_carrito: normalizarCarrito(c.numero_carrito),
      cedula: c.cedula ?? '',
      nombre_completo: c.nombre_completo ?? ''
    }))

    // Cargar cumplimiento diario
    const respuestaCumplimiento = await fetch(`${API_URL}/carritos/cumplimiento`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      }
    })

    if (respuestaCumplimiento.ok) {
      const dataCumplimiento = await respuestaCumplimiento.json() as RespuestaCumplimientoApi
      cumplimientoUsuarios.value = dataCumplimiento.usuarios ?? []
    }

    if (mantenerSeleccionId) {
      personalSeleccionado.value = personal.value.find((persona) => persona.id === mantenerSeleccionId) ?? null
    }
  } catch (error) {
    errorDatos.value = error instanceof Error ? error.message : 'No se pudo cargar la información'
  } finally {
    if (mostrarLoader) {
      cargandoDatos.value = false
    }
  }
}

function abrirGestion(p: UsuarioGestion) {
  personalSeleccionado.value = p
  busquedaCarritoModal.value = ''
  soloDisponibles.value = false
  mensajeAccionModal.value = ''
  tipoMensajeAccionModal.value = ''
  modalAbierto.value = true
}

function carritoEstaEnProceso(numeroCarrito: string): boolean {
  return cargandoAccionModal.value && carritoEnProceso.value === numeroCarrito
}

async function ejecutarGestionCarrito(numeroCarrito: string, endpoint: '/carritos/asignar' | '/carritos/quitar') {
  if (!personalSeleccionado.value || cargandoAccionModal.value) return

  cargandoAccionModal.value = true
  carritoEnProceso.value = numeroCarrito
  mensajeAccionModal.value = ''
  tipoMensajeAccionModal.value = ''

  try {
    const token = obtenerTokenSeguro()
    const numeroNormalizado = convertirNumeroCarrito(numeroCarrito)

    const respuesta = await fetch(`${API_URL}${endpoint}`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      },
      body: JSON.stringify({
        id_usuario: personalSeleccionado.value.id,
        numero_carrito: numeroNormalizado
      })
    })

    const data = await respuesta.json().catch(() => ({} as { mensaje?: string; error?: string }))
    if (!respuesta.ok) {
      throw new Error(data.error ?? 'No fue posible gestionar el carrito')
    }

    mensajeAccionModal.value = data.mensaje ?? 'Gestión aplicada correctamente'
    tipoMensajeAccionModal.value = 'ok'
    await cargarDatosGestion({ mostrarLoader: false, mantenerSeleccionId: personalSeleccionado.value.id })
  } catch (error) {
    mensajeAccionModal.value = error instanceof Error ? error.message : 'Error al gestionar el carrito'
    tipoMensajeAccionModal.value = 'error'
  } finally {
    cargandoAccionModal.value = false
    carritoEnProceso.value = ''
  }
}

async function asignarCarrito(numeroCarrito: string) {
  await ejecutarGestionCarrito(numeroCarrito, '/carritos/asignar')
}

async function quitarCarrito(numeroCarrito: string) {
  await ejecutarGestionCarrito(numeroCarrito, '/carritos/quitar')
}

async function transferirCarrito(numeroCarrito: string) {
  await ejecutarGestionCarrito(numeroCarrito, '/carritos/asignar')
}

async function crearUsuario() {
  if (!nuevoUsuario.value.nombre || !nuevoUsuario.value.empleado || !nuevoUsuario.value.cargo || !nuevoUsuario.value.email || !nuevoUsuario.value.clave) {
    mensajeCrearUsuario.value = 'Todos los campos son requeridos'
    tipoMensajeCrearUsuario.value = 'error'
    return
  }

  cargandoCrearUsuario.value = true
  mensajeCrearUsuario.value = ''
  tipoMensajeCrearUsuario.value = ''

  try {
    const token = obtenerTokenSeguro()
    const rolBackend = nuevoUsuario.value.rol === 'administrador' ? 'admin' : 'operario'

    const respuesta = await fetch(`${API_URL}/auth/usuarios`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      },
      body: JSON.stringify({
        empleado: nuevoUsuario.value.empleado,
        nombre_completo: nuevoUsuario.value.nombre,
        descripcion_cargo: nuevoUsuario.value.cargo,
        email: nuevoUsuario.value.email,
        contrasena: nuevoUsuario.value.clave,
        rol: rolBackend
      })
    })

    const data = await respuesta.json().catch(() => ({} as { mensaje?: string; error?: string }))
    if (!respuesta.ok) {
      throw new Error(data.error ?? 'No fue posible crear el usuario')
    }

    mensajeCrearUsuario.value = data.mensaje ?? 'Usuario creado correctamente'
    tipoMensajeCrearUsuario.value = 'ok'
    await cargarDatosGestion({ mostrarLoader: false })

    setTimeout(() => {
      modalCrearAbierto.value = false
      nuevoUsuario.value = { empleado: '', nombre: '', cargo: '', email: '', clave: '', rol: 'operario' }
      mensajeCrearUsuario.value = ''
      tipoMensajeCrearUsuario.value = ''
    }, 1200)
  } catch (error) {
    mensajeCrearUsuario.value = error instanceof Error ? error.message : 'Error al crear el usuario'
    tipoMensajeCrearUsuario.value = 'error'
  } finally {
    cargandoCrearUsuario.value = false
  }
}

function logout() {
  authStore.cerrarSesion()
  router.push('/login')
}

onMounted(() => {
  cargarDatosGestion()
})
</script>

<template>
  <div class="min-h-screen bg-white font-sans text-slate-800">
    
    <!-- Navbar (Mismo estilo que los otros dashboards) -->
    <header class="sticky top-0 z-30 w-full flex items-center justify-between px-7 py-4 bg-white/90 border-b border-slate-200 shadow-sm" style="backdrop-filter: blur(14px);">
      <div class="flex items-center gap-3">
        <img src="/favicon.ico" alt="SAO6" class="shrink-0" width="32" height="32" />
        <div class="flex flex-col leading-none gap-0.5">
          <span class="text-[0.82rem] font-extrabold text-slate-800 tracking-widest uppercase">ADMINISTRACIÓN</span>
          <span class="text-[0.6rem] font-semibold text-slate-500 tracking-widest uppercase">CONTROL DE PERSONAL</span>
        </div>
      </div>

      <div class="flex items-center gap-6">
        <div class="hidden sm:flex flex-col items-end leading-none gap-0.5 pl-4 border-l border-slate-200">
          <span class="text-[0.82rem] font-extrabold text-slate-800 tracking-wide uppercase">{{ usuarioLogueado?.nombre_completo || 'USUARIO' }}</span>
          <span class="text-[0.6rem] text-slate-500 tracking-widest uppercase">{{ usuarioLogueado?.rol || 'SIN ROL' }}</span>
        </div>
        <button @click="logout" class="flex items-center justify-center w-10 h-10 rounded-full bg-white border border-slate-200 text-slate-500 shadow-sm transition-all duration-200 hover:bg-red-50 hover:text-red-500 hover:border-red-200 cursor-pointer">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"></path><polyline points="16 17 21 12 16 7"></polyline><line x1="21" y1="12" x2="9" y2="12"></line></svg>
        </button>
      </div>
    </header>

    <main class="max-w-7xl mx-auto px-6 py-10">
      
      <!-- Encabezado con filtros y búsqueda -->
      <div class="flex flex-col md:flex-row md:items-end justify-between gap-6 mb-12">
        <div class="flex flex-col gap-2">
          <h1 class="text-4xl font-black text-slate-900 tracking-tight">Personal</h1>
          <p class="text-slate-500 font-medium">Gestiona y supervisa el inventario en tiempo real.</p>
        </div>

        <div class="flex flex-wrap items-center gap-4">
          <!-- Segmented Control de Estados -->
          <div class="flex bg-slate-100 p-1 rounded-2xl border border-slate-200 shadow-inner">
            <button 
              v-for="estado in ['Todos', 'Con carritos', 'Sin carritos']" 
              :key="estado"
              @click="filtroEstado = estado"
              :class="['px-5 py-2 text-xs font-black rounded-xl transition-all', filtroEstado === estado ? 'bg-white text-green-700 shadow-sm' : 'text-slate-500 hover:text-slate-700']"
            >
              {{ estado }}
            </button>
          </div>

          <!-- Barra de Búsqueda -->
          <div class="relative min-w-[300px]">
            <svg class="absolute left-4 top-1/2 -translate-y-1/2 text-slate-400" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><circle cx="11" cy="11" r="8"></circle><line x1="21" y1="21" x2="16.65" y2="16.65"></line></svg>
            <input 
              v-model="busqueda"
              type="text" 
              placeholder="Buscar usuario..." 
              class="w-full py-2.5 pl-11 pr-5 bg-white border-2 border-slate-100 rounded-2xl text-sm font-semibold focus:outline-none focus:border-green-400 transition-all shadow-sm"
            />
          </div>

          <!-- Botón Nueva Persona -->
          <button 
            @click="modalCrearAbierto = true"
            class="flex items-center gap-2 px-6 py-2.5 bg-green-500 text-white text-xs font-black rounded-2xl shadow-lg shadow-green-500/30 hover:bg-green-600 hover:-translate-y-0.5 transition-all cursor-pointer uppercase tracking-widest"
          >
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round"><line x1="12" y1="5" x2="12" y2="19"></line><line x1="5" y1="12" x2="19" y2="12"></line></svg>
            Nueva Persona
          </button>
        </div>
      </div>

      <div v-if="errorDatos" class="mb-8 rounded-2xl border border-red-200 bg-red-50 px-5 py-4 text-sm font-bold text-red-700">
        {{ errorDatos }}
      </div>

      <div v-if="cargandoDatos" class="mb-8 rounded-2xl border border-slate-200 bg-slate-50 px-5 py-4 text-sm font-bold text-slate-600 flex items-center gap-3">
        <span class="inline-block h-4 w-4 animate-spin rounded-full border-2 border-slate-300 border-t-green-500"></span>
        Cargando personal y carritos...
      </div>

      <!-- Grid de Cards de Personal -->
      <TransitionGroup
        v-if="!cargandoDatos"
        name="staggered-fade" 
        tag="div" 
        class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8"
      >
        <div 
          v-for="(p, index) in personalFiltrado" 
          :key="p.id"
          @click="abrirGestion(p)"
          class="card-admin relative bg-white border border-slate-100 rounded-[2.5rem] p-8 shadow-xl shadow-slate-200/40 hover:shadow-2xl hover:-translate-y-1 transition-all duration-300 group overflow-hidden cursor-pointer"
          :style="`animation-delay: ${index * 0.05}s`"
        >
          <!-- Botón de ajustes superior derecho -->
          <button class="absolute top-6 right-6 p-2 text-slate-300 hover:text-slate-500 transition-colors">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><circle cx="12" cy="12" r="3"></circle><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z"></path></svg>
          </button>

          <!-- Información Principal -->
          <div class="mb-8">
            <h3 class="text-xl font-black text-slate-800 leading-tight mb-2 pr-8">{{ p.nombre }}</h3>
            <div class="flex items-center gap-2 text-xs font-bold text-slate-400">
              <span class="w-1.5 h-1.5 rounded-full bg-green-500"></span>
              {{ p.email }}
            </div>
          </div>

          <div class="flex flex-col gap-3 mb-8">
            <div class="bg-slate-50/50 rounded-3xl p-5 flex items-center justify-between border border-slate-100">
              <div class="flex flex-col">
                <span class="text-3xl font-black text-slate-800 leading-none mb-1">{{ p.carritos.length }}</span>
                <span class="text-[0.6rem] font-black text-slate-400 uppercase tracking-widest">CARRITOS</span>
              </div>
              <button class="text-[0.65rem] font-black uppercase tracking-widest text-green-600 hover:text-green-700 cursor-pointer">
                Ver gestión
              </button>
            </div>

            <div class="flex flex-wrap gap-2 min-h-8">
              <span v-if="p.carritos.length === 0" class="text-[0.65rem] font-black uppercase tracking-wider text-slate-400">Sin carritos asignados</span>
              <span v-for="carrito in p.carritos.slice(0, 5)" :key="`${p.id}-${carrito}`" class="px-2.5 py-1 rounded-full bg-green-50 text-green-700 border border-green-100 text-[0.65rem] font-black">
                {{ carrito }}
              </span>
              <span v-if="p.carritos.length > 5" class="px-2.5 py-1 rounded-full bg-slate-100 text-slate-600 border border-slate-200 text-[0.65rem] font-black">
                +{{ p.carritos.length - 5 }}
              </span>
            </div>
          </div>

          <!-- Cumplimiento Diario -->
          <div class="pt-5 border-t border-slate-100">
            <div class="flex items-center justify-between mb-2">
              <span class="text-[0.65rem] font-black text-slate-400 uppercase tracking-widest">CUMPLIMIENTO HOY</span>
              <span v-if="obtenerCumplimientoUsuario(p.id)" class="text-[0.75rem] font-black"
                :class="{
                  'text-green-600': obtenerCumplimientoUsuario(p.id)!.porcentaje_global >= 100,
                  'text-yellow-600': obtenerCumplimientoUsuario(p.id)!.porcentaje_global > 0 && obtenerCumplimientoUsuario(p.id)!.porcentaje_global < 100,
                  'text-slate-400': obtenerCumplimientoUsuario(p.id)!.porcentaje_global === 0
                }">
                {{ obtenerCumplimientoUsuario(p.id)!.porcentaje_global }}%
              </span>
              <span v-else class="text-[0.75rem] font-black text-slate-400">--</span>
            </div>

            <!-- Barra de progreso -->
            <div class="relative h-2.5 bg-slate-100 rounded-full overflow-hidden">
              <div v-if="obtenerCumplimientoUsuario(p.id)"
                class="h-full rounded-full transition-all duration-500"
                :class="{
                  'bg-green-500': obtenerCumplimientoUsuario(p.id)!.porcentaje_global >= 100,
                  'bg-yellow-500': obtenerCumplimientoUsuario(p.id)!.porcentaje_global > 0 && obtenerCumplimientoUsuario(p.id)!.porcentaje_global < 100,
                  'bg-slate-300': obtenerCumplimientoUsuario(p.id)!.porcentaje_global === 0
                }"
                :style="{ width: Math.min(obtenerCumplimientoUsuario(p.id)!.porcentaje_global, 100) + '%' }">
              </div>
              <div v-else class="h-full bg-slate-200 rounded-full" style="width: 0%"></div>
            </div>

            <!-- Detalle de completados -->
            <div v-if="obtenerCumplimientoUsuario(p.id) && obtenerCumplimientoUsuario(p.id)!.total_registros > 0" class="flex items-center justify-between mt-2">
              <span class="text-[0.6rem] font-bold text-slate-400">
                {{ obtenerCumplimientoUsuario(p.id)!.total_completados }} de {{ obtenerCumplimientoUsuario(p.id)!.total_registros }} ítems
              </span>
              <span class="text-[0.6rem] font-black uppercase tracking-wider"
                :class="{
                  'text-green-600': obtenerCumplimientoUsuario(p.id)!.estado === 'completado',
                  'text-yellow-600': obtenerCumplimientoUsuario(p.id)!.estado === 'en_progreso',
                  'text-red-500': obtenerCumplimientoUsuario(p.id)!.estado === 'pendiente',
                  'text-slate-400': obtenerCumplimientoUsuario(p.id)!.estado === 'sin_registros'
                }">
                {{ obtenerCumplimientoUsuario(p.id)!.estado === 'completado' ? 'Completado' :
                   obtenerCumplimientoUsuario(p.id)!.estado === 'en_progreso' ? 'En progreso' :
                   obtenerCumplimientoUsuario(p.id)!.estado === 'pendiente' ? 'Pendiente' : 'Sin registros' }}
              </span>
            </div>
            <div v-else-if="p.carritos.length > 0" class="mt-2">
              <span class="text-[0.6rem] font-bold text-slate-400">Sin registros para hoy</span>
            </div>
            <div v-else class="mt-2">
              <span class="text-[0.6rem] font-bold text-slate-400">Sin carritos asignados</span>
            </div>
          </div>
        </div>
      </TransitionGroup>

      <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8">
        <div v-for="n in 6" :key="`loader-card-${n}`" class="rounded-[2.5rem] border border-slate-100 bg-white p-8 shadow-xl shadow-slate-200/30 animate-pulse">
          <div class="h-6 w-3/4 rounded-xl bg-slate-100 mb-4"></div>
          <div class="h-4 w-2/3 rounded-xl bg-slate-100 mb-8"></div>
          <div class="h-16 w-full rounded-3xl bg-slate-100 mb-6"></div>
          <div class="h-4 w-1/2 rounded-xl bg-slate-100"></div>
        </div>
      </div>

      <!-- Estado Vacío -->
      <div v-if="!cargandoDatos && personalFiltrado.length === 0" class="flex flex-col items-center justify-center py-24 text-center">
        <div class="w-20 h-20 bg-slate-100 rounded-3xl flex items-center justify-center mb-6">
          <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="#cbd5e1" stroke-width="2"><circle cx="11" cy="11" r="8"></circle><line x1="21" y1="21" x2="16.65" y2="16.65"></line></svg>
        </div>
        <h3 class="text-xl font-black text-slate-800 mb-2">No se encontraron usuarios</h3>
        <p class="text-slate-400 font-medium">Prueba con otro nombre o ajusta los filtros de estado.</p>
      </div>

      <!-- Modal de Gestión de Usuario -->
      <Teleport to="body">
        <Transition
          enter-active-class="transition duration-300 ease-out" enter-from-class="opacity-0" enter-to-class="opacity-100"
          leave-active-class="transition duration-200 ease-in" leave-from-class="opacity-100" leave-to-class="opacity-0"
        >
          <div v-if="modalAbierto" class="fixed inset-0 z-[100] flex items-center justify-center px-4">
            <!-- Backdrop -->
            <div class="absolute inset-0 bg-slate-900/60 backdrop-blur-sm shadow-sm" @click="modalAbierto = false"></div>

            <!-- Contenido Modal -->
            <Transition
              appear
              enter-active-class="transition-all duration-500 ease-[cubic-bezier(0.34,1.56,0.64,1)] transform"
              enter-from-class="opacity-0 scale-90 translate-y-12"
              enter-to-class="opacity-100 scale-100 translate-y-0"
            >
              <div class="relative bg-white w-full max-w-6xl rounded-[3rem] shadow-2xl overflow-hidden flex flex-col" style="height: 750px;">
                
                <!-- Modal Header -->
                <div class="px-12 py-8 border-b border-slate-100 flex items-center justify-between bg-white">
                  <div class="flex flex-col gap-1.5">
                    <h2 class="text-3xl font-black text-slate-800 tracking-tight">Gestión de Usuario</h2>
                    <p class="text-[0.7rem] font-bold tracking-[0.15em] text-slate-400 uppercase">
                      Editando a <span class="text-green-500">{{ personalSeleccionado?.nombre }}</span>
                    </p>
                  </div>
                  <button @click="modalAbierto = false" class="w-12 h-12 flex items-center justify-center rounded-2xl bg-slate-50 text-slate-400 hover:bg-red-50 hover:text-red-500 transition-all cursor-pointer border border-slate-100/50">
                    <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><line x1="18" y1="6" x2="6" y2="18"></line><line x1="6" y1="6" x2="18" y2="18"></line></svg>
                  </button>
                </div>

                <!-- Modal Body (Layout Lateral) -->
                <div class="flex-1 flex overflow-hidden">
                  <!-- Sidebar Modal -->
                  <aside class="w-72 border-r border-slate-50 bg-slate-50/40 p-8 flex flex-col gap-3">
                    <button 
                      @click="tabActivo = 'Carritos'"
                      :class="['flex items-center gap-4 px-6 py-4 rounded-2xl text-[0.8rem] font-black uppercase tracking-widest transition-all cursor-pointer border', 
                        tabActivo === 'Carritos' ? 'bg-green-500 text-white shadow-xl shadow-green-500/30 border-green-400' : 'text-slate-400 bg-white/50 border-transparent hover:bg-white hover:text-slate-600 hover:shadow-sm']"
                    >
                      <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><rect x="2" y="7" width="20" height="14" rx="2" ry="2"></rect><path d="M16 21V5a2 2 0 0 0-2-2h-4a2 2 0 0 0-2 2v16"></path></svg>
                      Carritos
                    </button>
                    <button 
                      @click="tabActivo = 'Movimientos'"
                      :class="['flex items-center gap-4 px-6 py-4 rounded-2xl text-[0.8rem] font-black uppercase tracking-widest transition-all cursor-pointer border', 
                        tabActivo === 'Movimientos' ? 'bg-green-500 text-white shadow-xl shadow-green-500/30 border-green-400' : 'text-slate-400 bg-white/50 border-transparent hover:bg-white hover:text-slate-600 hover:shadow-sm']"
                    >
                      <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><circle cx="12" cy="12" r="10"></circle><polyline points="12 6 12 12 16 14"></polyline></svg>
                      Movimientos
                    </button>
                  </aside>

                  <!-- Content Area -->
                  <main class="flex-1 p-10 overflow-y-auto bg-slate-50/10">
                    <div v-if="tabActivo === 'Carritos'" class="flex flex-col gap-8">
                      <div v-if="cargandoAccionModal" class="rounded-2xl border border-slate-200 bg-slate-50 px-4 py-3 text-sm font-bold text-slate-600 flex items-center gap-3">
                        <span class="inline-block h-4 w-4 animate-spin rounded-full border-2 border-slate-300 border-t-green-500"></span>
                        Procesando gestión de carritos...
                      </div>

                      <div
                        v-if="mensajeAccionModal"
                        class="rounded-2xl px-4 py-3 text-sm font-bold border"
                        :class="tipoMensajeAccionModal === 'ok' ? 'bg-green-50 border-green-200 text-green-700' : 'bg-red-50 border-red-200 text-red-700'"
                      >
                        {{ mensajeAccionModal }}
                      </div>

                      <div class="flex items-center justify-between">
                        <div class="flex flex-col gap-1">
                          <h3 class="text-sm font-black text-slate-800 uppercase tracking-[0.2em]">Carritos de la Persona</h3>
                          <p class="text-[0.65rem] font-bold text-slate-400 tracking-wider">GESTIONA LA DISTRIBUCIÓN DE HERRAMIENTAS</p>
                        </div>
                        <span class="text-[0.7rem] font-black bg-slate-100 text-slate-500 px-4 py-2 rounded-full uppercase tracking-wider border border-slate-200/50">{{ carritosUsuarioSeleccionado.length }} ASIGNADOS</span>
                      </div>

                      <div class="flex flex-wrap gap-2 min-h-10">
                        <span v-if="carritosUsuarioSeleccionado.length === 0" class="text-[0.75rem] font-black text-slate-400 uppercase tracking-widest">Sin carritos asignados</span>
                        <div v-for="numero in carritosUsuarioSeleccionado" :key="`sel-${numero}`" class="flex items-center gap-2 rounded-full border border-green-100 bg-green-50 px-3 py-1.5">
                          <span class="text-[0.7rem] font-black text-green-700">{{ numero }}</span>
                          <button @click="quitarCarrito(numero)" :disabled="carritoEstaEnProceso(numero)" class="w-5 h-5 rounded-full bg-white text-red-500 hover:bg-red-50 border border-red-100 text-[0.65rem] font-black cursor-pointer disabled:opacity-60 disabled:cursor-not-allowed">
                            ×
                          </button>
                        </div>
                      </div>

                      <!-- Modal Search -->
                      <div class="flex flex-col gap-4">
                        <div class="relative group">
                          <svg class="absolute left-5 top-1/2 -translate-y-1/2 text-slate-300 group-focus-within:text-green-500 transition-colors" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><circle cx="11" cy="11" r="8"></circle><line x1="21" y1="21" x2="16.65" y2="16.65"></line></svg>
                          <input v-model="busquedaCarritoModal" type="text" placeholder="Buscar por número, cédula o ubicación..." class="w-full bg-white border border-slate-200 py-4 pl-14 pr-6 rounded-[1.5rem] text-sm font-bold text-slate-700 focus:outline-none focus:border-green-400 focus:ring-4 focus:ring-green-50/50 transition-all shadow-sm" />
                        </div>
                        <label class="flex items-center gap-3 cursor-pointer group w-fit ml-2">
                          <input v-model="soloDisponibles" type="checkbox" class="w-5 h-5 rounded-lg border-slate-200 text-green-500 focus:ring-green-400 cursor-pointer" />
                          <span class="text-[0.75rem] font-black text-slate-400 group-hover:text-slate-600 transition-colors uppercase tracking-wider">Mostrar solo disponibles</span>
                        </label>
                      </div>

                      <!-- Listado de carritos -->
                      <div class="flex flex-col gap-5">
                        <div v-for="c in carritosCatalogoFiltrados" :key="c.numero_carrito" class="bg-white border border-slate-100 rounded-[2rem] p-6 flex items-center gap-7 shadow-sm hover:shadow-xl hover:-translate-y-0.5 transition-all group">
                          <div class="w-16 h-16 rounded-[1.25rem] bg-orange-50 border border-orange-100 flex items-center justify-center shrink-0 group-hover:bg-orange-100 transition-colors">
                            <svg width="26" height="26" viewBox="0 0 24 24" fill="none" stroke="#f97316" stroke-width="2"><rect x="2" y="7" width="20" height="14" rx="2" ry="2"></rect><path d="M16 21V5a2 2 0 0 0-2-2h-4a2 2 0 0 0-2 2v16"></path></svg>
                          </div>
                          <div class="flex-1 min-w-0">
                            <div class="flex items-center gap-3 mb-2">
                              <span class="text-[0.85rem] font-black text-slate-800 uppercase tracking-tight">Carrito {{ c.numero_carrito }}</span>
                              <div class="flex items-center gap-2">
                                <span class="w-1 h-1 rounded-full bg-slate-300"></span>
                                <span class="text-[0.65rem] font-black text-slate-400 tracking-wider">CÉDULA: {{ c.cedula || 'N/A' }}</span>
                                <span class="text-[0.65rem] font-black px-2.5 py-0.5 rounded-full border"
                                  :class="asignacionPorCarrito.get(c.numero_carrito) ? 'text-orange-600 bg-orange-50 border-orange-100' : 'text-green-600 bg-green-50 border-green-100'">
                                  {{ asignacionPorCarrito.get(c.numero_carrito) ? 'EN USO' : 'DISPONIBLE' }}
                                </span>
                              </div>
                            </div>
                            <p class="text-[0.75rem] font-black text-slate-600 mb-2 truncate bg-slate-50/80 px-3 py-1 rounded-lg w-fit border border-slate-100">{{ c.nombre_completo || 'SIN UBICACIÓN' }}</p>
                            <div class="flex items-center gap-2 text-[0.65rem] font-black uppercase tracking-[0.1em]"
                              :class="asignacionPorCarrito.get(c.numero_carrito) ? 'text-orange-400' : 'text-green-500'">
                              <div class="w-5 h-5 rounded-full bg-orange-100 flex items-center justify-center">
                                <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3.5"><path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"></path><circle cx="12" cy="7" r="4"></circle></svg>
                              </div>
                              <template v-if="asignacionPorCarrito.get(c.numero_carrito)">
                                En uso por:
                                <span class="text-orange-600">{{ asignacionPorCarrito.get(c.numero_carrito)?.nombre }}</span>
                              </template>
                              <template v-else>
                                Sin asignar
                              </template>
                            </div>
                          </div>
                          <div class="flex items-center gap-2">
                            <button
                              v-if="!asignacionPorCarrito.get(c.numero_carrito)"
                              @click="asignarCarrito(c.numero_carrito)"
                              :disabled="cargandoAccionModal"
                              class="px-4 py-2 rounded-xl bg-green-500 text-white text-[0.65rem] font-black uppercase tracking-wider hover:bg-green-600 cursor-pointer disabled:opacity-60 disabled:cursor-not-allowed"
                            >
                              {{ carritoEstaEnProceso(c.numero_carrito) ? 'Asignando...' : 'Asignar' }}
                            </button>
                            <button
                              v-else-if="asignacionPorCarrito.get(c.numero_carrito)?.id !== personalSeleccionado?.id"
                              @click="transferirCarrito(c.numero_carrito)"
                              :disabled="cargandoAccionModal"
                              class="px-4 py-2 rounded-xl bg-orange-500 text-white text-[0.65rem] font-black uppercase tracking-wider hover:bg-orange-600 cursor-pointer disabled:opacity-60 disabled:cursor-not-allowed"
                            >
                              {{ carritoEstaEnProceso(c.numero_carrito) ? 'Transfiriendo...' : 'Transferir' }}
                            </button>
                            <button
                              v-else
                              @click="quitarCarrito(c.numero_carrito)"
                              :disabled="cargandoAccionModal"
                              class="px-4 py-2 rounded-xl bg-red-500 text-white text-[0.65rem] font-black uppercase tracking-wider hover:bg-red-600 cursor-pointer disabled:opacity-60 disabled:cursor-not-allowed"
                            >
                              {{ carritoEstaEnProceso(c.numero_carrito) ? 'Quitando...' : 'Quitar' }}
                            </button>
                          </div>
                        </div>
                      </div>
                    </div>
                    
                    <div v-else class="flex flex-col items-center justify-center py-20 text-center text-slate-300">
                      <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><circle cx="12" cy="12" r="10"></circle><polyline points="12 6 12 12 16 14"></polyline></svg>
                      <p class="mt-4 text-sm font-black uppercase tracking-widest">Historial de Movimientos</p>
                      <p class="text-xs font-medium">No hay registros recientes para este usuario.</p>
                    </div>
                  </main>
                </div>

                <!-- Modal Footer -->
                <div class="px-10 py-6 bg-slate-50 border-t border-slate-100 flex justify-center">
                  <button @click="modalAbierto = false" class="w-full max-w-lg py-4 bg-slate-900 text-white rounded-2xl text-[0.75rem] font-black uppercase tracking-[0.3em] shadow-xl shadow-slate-900/20 hover:bg-slate-800 transition-all cursor-pointer">
                    Guardar y Cerrar
                  </button>
                </div>

              </div>
            </Transition>
          </div>
        </Transition>
      </Teleport>

      <!-- Modal Crear Nueva Persona -->
      <Teleport to="body">
        <Transition
          enter-active-class="transition duration-300 ease-out" enter-from-class="opacity-0" enter-to-class="opacity-100"
          leave-active-class="transition duration-200 ease-in" leave-from-class="opacity-100" leave-to-class="opacity-0"
        >
          <div v-if="modalCrearAbierto" class="fixed inset-0 z-[110] flex items-center justify-center px-4">
            <div class="absolute inset-0 bg-slate-900/40 backdrop-blur-sm shadow-sm" @click="modalCrearAbierto = false"></div>

            <Transition
              appear
              enter-active-class="transition-all duration-500 ease-[cubic-bezier(0.34,1.56,0.64,1)] transform"
              enter-from-class="opacity-0 scale-90 translate-y-12"
              enter-to-class="opacity-100 scale-100 translate-y-0"
            >
              <div class="relative bg-white w-full max-w-xl rounded-[2.5rem] shadow-2xl overflow-hidden p-10 flex flex-col gap-8">
                <div class="flex items-center justify-between">
                  <div class="flex flex-col gap-1">
                    <h2 class="text-2xl font-black text-slate-800 tracking-tight">Registro de Personal</h2>
                    <p class="text-[0.65rem] font-bold text-slate-400 uppercase tracking-widest">Ingresa los datos del nuevo integrante</p>
                  </div>
                  <button @click="modalCrearAbierto = false" class="text-slate-300 hover:text-red-500 transition-colors">
                    <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><line x1="18" y1="6" x2="6" y2="18"></line><line x1="6" y1="6" x2="18" y2="18"></line></svg>
                  </button>
                </div>

                <div v-if="mensajeCrearUsuario" class="rounded-2xl px-4 py-3 text-sm font-bold border" :class="tipoMensajeCrearUsuario === 'ok' ? 'bg-green-50 border-green-200 text-green-700' : 'bg-red-50 border-red-200 text-red-700'">
                  {{ mensajeCrearUsuario }}
                </div>

                <div class="grid grid-cols-1 gap-5">
                  <div class="flex flex-col gap-1.5">
                    <label class="text-[0.62rem] font-black text-slate-400 uppercase tracking-widest pl-1">Nombre Completo</label>
                    <input v-model="nuevoUsuario.nombre" type="text" placeholder="Ej: JUAN PEREZ" class="w-full bg-slate-50 border-2 border-slate-100 py-3 px-5 rounded-2xl text-sm font-bold text-slate-700 focus:outline-none focus:border-green-400 focus:bg-white transition-all shadow-sm" />
                  </div>

                  <div class="grid grid-cols-2 gap-5">
                    <div class="flex flex-col gap-1.5">
                      <label class="text-[0.62rem] font-black text-slate-400 uppercase tracking-widest pl-1">Número de Empleado</label>
                      <input v-model="nuevoUsuario.empleado" type="text" placeholder="Ej: EMP001" class="w-full bg-slate-50 border-2 border-slate-100 py-3 px-5 rounded-2xl text-sm font-bold text-slate-700 focus:outline-none focus:border-green-400 focus:bg-white transition-all shadow-sm" />
                    </div>
                    <div class="flex flex-col gap-1.5">
                      <label class="text-[0.62rem] font-black text-slate-400 uppercase tracking-widest pl-1">Cargo</label>
                      <input v-model="nuevoUsuario.cargo" type="text" placeholder="Ej: OPERARIO" class="w-full bg-slate-50 border-2 border-slate-100 py-3 px-5 rounded-2xl text-sm font-bold text-slate-700 focus:outline-none focus:border-green-400 focus:bg-white transition-all shadow-sm" />
                    </div>
                    <div class="flex flex-col gap-1.5">
                      <label class="text-[0.62rem] font-black text-slate-400 uppercase tracking-widest pl-1">Rol de Usuario</label>
                      <select v-model="nuevoUsuario.rol" class="w-full bg-slate-50 border-2 border-slate-100 py-3 px-5 rounded-2xl text-sm font-bold text-slate-700 focus:outline-none focus:border-green-400 focus:bg-white transition-all shadow-sm appearance-none">
                        <option value="operario">Operario</option>
                        <option value="administrador">Administrador</option>
                      </select>
                    </div>
                  </div>

                  <div class="flex flex-col gap-1.5">
                    <label class="text-[0.62rem] font-black text-slate-400 uppercase tracking-widest pl-1">Correo Electrónico</label>
                    <input v-model="nuevoUsuario.email" type="email" placeholder="ejemplo@sao6.com.co" class="w-full bg-slate-50 border-2 border-slate-100 py-3 px-5 rounded-2xl text-sm font-bold text-slate-700 focus:outline-none focus:border-green-400 focus:bg-white transition-all shadow-sm" />
                  </div>

                  <div class="flex flex-col gap-1.5">
                    <label class="text-[0.62rem] font-black text-slate-400 uppercase tracking-widest pl-1">Contraseña Inicial</label>
                    <div class="relative group">
                      <input v-model="nuevoUsuario.clave" type="password" placeholder="********" class="w-full bg-slate-50 border-2 border-slate-100 py-3 px-5 rounded-2xl text-sm font-bold text-slate-700 focus:outline-none focus:border-green-400 focus:bg-white transition-all shadow-sm" />
                      <div class="absolute right-5 top-1/2 -translate-y-1/2 text-slate-300">
                        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><path d="M1 12s4-8 11-8 11-8 11-8-4 8-11 8-11-8-11-8z"></path><circle cx="12" cy="12" r="3"></circle></svg>
                      </div>
                    </div>
                  </div>
                </div>

                <button @click="crearUsuario" :disabled="cargandoCrearUsuario" class="w-full py-4 bg-green-500 text-white rounded-2xl text-sm font-black uppercase tracking-[0.2em] shadow-xl shadow-green-500/20 hover:bg-green-600 hover:-translate-y-1 transition-all cursor-pointer disabled:opacity-60 disabled:cursor-not-allowed flex items-center justify-center gap-3">
                  <span v-if="cargandoCrearUsuario" class="inline-block h-4 w-4 animate-spin rounded-full border-2 border-white border-t-transparent"></span>
                  {{ cargandoCrearUsuario ? 'Creando usuario...' : 'Finalizar Registro' }}
                </button>
              </div>
            </Transition>
          </div>
        </Transition>
      </Teleport>

    </main>
  </div>
</template>

<style scoped>
.card-admin {
  animation: card-in 0.6s cubic-bezier(0.22, 1, 0.36, 1) both;
}

@keyframes card-in {
  from { opacity: 0; transform: translateY(30px) scale(0.95); }
  to { opacity: 1; transform: translateY(0) scale(1); }
}

.staggered-fade-enter-active { transition: all 0.5s cubic-bezier(0.34, 1.56, 0.64, 1); }
.staggered-fade-leave-active { transition: all 0.3s ease-in; position: absolute; }
.staggered-fade-enter-from, .staggered-fade-leave-to { opacity: 0; transform: translateY(20px) scale(0.9); }
.staggered-fade-move { transition: transform 0.5s cubic-bezier(0.34, 1.56, 0.64, 1); }

/* Personalización del scroll si es necesario */
::-webkit-scrollbar { width: 6px; }
::-webkit-scrollbar-track { background: transparent; }
::-webkit-scrollbar-thumb { background: #e2e8f0; border-radius: 10px; }
::-webkit-scrollbar-thumb:hover { background: #cbd5e1; }
</style>
