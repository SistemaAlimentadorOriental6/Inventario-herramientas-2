<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../../auth/store/useAuthStore'

const router = useRouter()
const authStore = useAuthStore()

const usuarioLogueado = computed(() => authStore.usuario)
const esOperario = computed(() => usuarioLogueado.value?.rol === 'operario')

const cargando = ref(true)
const errorCarga = ref('')
const mensajeExito = ref('')
const mensajeError = ref('')

// Lista de herramientas desde el endpoint
const herramientas = ref<any[]>([])

const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080'

function obtenerTokenSeguro(): string {
  const token = authStore.token
  if (!token) {
    throw new Error('No hay sesión activa')
  }
  return token
}

async function cargarItemsPrestamo() {
  cargando.value = true
  errorCarga.value = ''
  try {
    const token = obtenerTokenSeguro()
    const respuesta = await fetch(`${API_URL}/prestamo/items`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      }
    })

    if (!respuesta.ok) {
      const error = await respuesta.json().catch(() => ({}))
      throw new Error(error.error ?? 'Error al cargar ítems de préstamo')
    }

    const data = await respuesta.json()
    // Mapear datos del endpoint a la estructura de la UI
    herramientas.value = (data.items || []).map((item: any, index: number) => ({
      id: index + 1,
      codigo: item.f_referencia || '',
      nombre: item.f_desc_item || '',
      marca: item.f121_id_ext1_detalle || '',
      categoria: 'General',
      disponible: (item.f_cant_existencia_1 || 0) > 0,
      stock: item.f_cant_existencia_1 || 0,
      um: item.f_um || '',
      prestadosA: []
    }))

    // Cargar préstamos activos y asociarlos a las herramientas
    await cargarPrestamosActivos()
  } catch (error) {
    errorCarga.value = error instanceof Error ? error.message : 'Error al cargar datos'
  } finally {
    cargando.value = false
  }
}

async function cargarPrestamosActivos() {
  try {
    const token = obtenerTokenSeguro()
    const respuesta = await fetch(`${API_URL}/prestamos?estado=activo`, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      }
    })

    if (!respuesta.ok) {
      console.error('Error al cargar préstamos activos')
      return
    }

    const data = await respuesta.json()
    const prestamos = data.prestamos || []

    // Asociar préstamos a las herramientas
    prestamos.forEach((prestamo: any) => {
      const herramienta = herramientas.value.find(h => h.codigo === prestamo.referencia)
      if (herramienta) {
        herramienta.prestadosA.push({
          id_prestamo: prestamo.id_prestamo,
          id: prestamo.cedula_operario,
          nombre: prestamo.nombre_operario,
          cantidad: prestamo.cantidad_prestada,
          fecha_prestamo: prestamo.fecha_prestamo
        })
        // Actualizar stock disponible
        herramienta.stock = Math.max(0, herramienta.stock - prestamo.cantidad_prestada)
        if (herramienta.stock <= 0) {
          herramienta.disponible = false
        }
      }
    })
  } catch (error) {
    console.error('Error al cargar préstamos activos:', error)
  }
}

onMounted(() => {
  cargarItemsPrestamo()
})

const busqueda = ref('')
const filtroMarca = ref('Todas')
const filtroEstado = ref('Todos') // Todos, Disponibles, Prestados
const ordenStock = ref('default') // default, mayor-menor, menor-mayor

// Listas dinámicas para los filtros
const marcasUnicas = computed(() => {
  const m = ['Todas', ...new Set(herramientas.value.map(h => h.marca))]
  return m.sort()
})

// Función para normalizar texto (quitar tildes y acentos)
function normalizar(texto: string) {
  return texto?.normalize("NFD").replace(/[\u0300-\u036f]/g, "").toLowerCase() || ""
}

// Computed properties para facilitar el filtro y orden
const herramientasFiltradas = computed(() => {
  let filtradas = [...herramientas.value]

  // 1. Filtro por búsqueda de texto
  if (busqueda.value) {
    const query = normalizar(busqueda.value)
    filtradas = filtradas.filter(h => 
      normalizar(h.nombre).includes(query) || 
      normalizar(h.codigo).includes(query) ||
      normalizar(h.marca).includes(query)
    )
  }

  // 2. Filtro por Marca
  if (filtroMarca.value !== 'Todas') {
    filtradas = filtradas.filter(h => h.marca === filtroMarca.value)
  }

  // 3. Filtro por Estado (Disponibilidad / Préstamo)
  if (filtroEstado.value === 'Disponibles') {
    filtradas = filtradas.filter(h => h.disponible)
  } else if (filtroEstado.value === 'Prestados') {
    filtradas = filtradas.filter(h => h.prestadosA.length > 0)
  }

  // 5. Ordenamiento por Stock
  if (ordenStock.value === 'mayor-menor') {
    filtradas.sort((a, b) => b.stock - a.stock)
  } else if (ordenStock.value === 'menor-mayor') {
    filtradas.sort((a, b) => a.stock - b.stock)
  }

  return filtradas
})

const modalAbierto = ref(false)
const modalDevolucionAbierto = ref(false)
const herramientaSeleccionada = ref<any>(null)
const personaSeleccionada = ref<any>(null)
const procesandoID = ref<number | null>(null) // Para manejar el loader por card

// Estados para los dropdowns personalizados
const menuAbierto = ref<string | null>(null) // 'marca', 'categoria', 'orden'

function toggleMenu(nombre: string) {
  if (menuAbierto.value === nombre) menuAbierto.value = null
  else menuAbierto.value = nombre
}

const busquedaPersonal = ref('')
const cargandoPersonal = ref(false)
const errorPersonal = ref('')

// Lista de personal desde vw_Ubicaciones
const personal = ref<any[]>([])

// Extensiones a probar para las fotos de empleados
const EXTENSIONES_FOTO = ['jpg', 'jpeg', 'png']
const BASE_URL_FOTOS = 'http://admon.sao6.com.co/web/uploads/empleados'

// Construir URL de foto con extensión por defecto jpg
function construirUrlFoto(cedula: string): string {
  return `${BASE_URL_FOTOS}/${cedula}.jpg`
}

// Manejar error de carga de imagen - intentar siguiente extensión o fallback a iniciales
function manejarErrorFoto(event: Event, persona: any) {
  const img = event.target as HTMLImageElement
  const cedula = persona.id
  const urlActual = img.src

  // Extraer extensión actual
  const match = urlActual.match(/\.([^.]+)$/)
  const extActual = match ? match[1].toLowerCase() : 'jpg'
  const indiceActual = EXTENSIONES_FOTO.indexOf(extActual)

  // Intentar siguiente extensión
  if (indiceActual >= 0 && indiceActual < EXTENSIONES_FOTO.length - 1) {
    const siguienteExt = EXTENSIONES_FOTO[indiceActual + 1]
    img.src = `${BASE_URL_FOTOS}/${cedula}.${siguienteExt}`
  } else {
    // Fallback a dicebear con iniciales
    const iniciales = persona.nombre.substring(0, 2).toUpperCase()
    img.src = `https://api.dicebear.com/7.x/initials/svg?seed=${iniciales}`
  }
}

async function cargarPersonal() {
  cargandoPersonal.value = true
  errorPersonal.value = ''
  try {
    const token = obtenerTokenSeguro()

    // Cargar personal y préstamos activos en paralelo
    const [respPersonal, respPrestamos] = await Promise.all([
      fetch(`${API_URL}/carritos/usuarios-ubicacion`, {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`
        }
      }),
      fetch(`${API_URL}/prestamos?estado=activo`, {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`
        }
      })
    ])

    if (!respPersonal.ok) {
      const error = await respPersonal.json().catch(() => ({}))
      throw new Error(error.error ?? 'Error al cargar personal')
    }

    const dataPersonal = await respPersonal.json()
    const usuarios = dataPersonal.usuarios || []

    // Obtener conteo de préstamos por operario
    let prestamosPorOperario: Record<string, number> = {}
    if (respPrestamos.ok) {
      const dataPrestamos = await respPrestamos.json()
      const prestamos = dataPrestamos.prestamos || []
      prestamosPorOperario = prestamos.reduce((acc: Record<string, number>, p: any) => {
        const cedula = p.cedula_operario || p.cedulaOperario
        if (cedula) {
          acc[cedula] = (acc[cedula] || 0) + (p.cantidad_prestada || 1)
        }
        return acc
      }, {})
    }

    // Mapear datos del endpoint a la estructura de la UI
    personal.value = usuarios.map((u: any) => ({
      id: u.cedula || '',
      nombre: u.nombre || '',
      rol: 'OPERARIO',
      foto: construirUrlFoto(u.cedula || ''),
      fotoFallback: `https://api.dicebear.com/7.x/initials/svg?seed=${(u.nombre || '').substring(0, 2).toUpperCase()}`,
      prestamosActivos: prestamosPorOperario[u.cedula] || 0
    }))

    // Ordenar: primero los que tienen préstamos activos
    personal.value.sort((a, b) => b.prestamosActivos - a.prestamosActivos)
  } catch (error) {
    errorPersonal.value = error instanceof Error ? error.message : 'Error al cargar personal'
  } finally {
    cargandoPersonal.value = false
  }
}

const personalFiltrado = computed(() => {
  if (!busquedaPersonal.value) return personal.value
  const q = busquedaPersonal.value.toLowerCase()
  return personal.value.filter(p => p.nombre.toLowerCase().includes(q) || String(p.id).includes(q))
})

function solicitarPrestamo(herramienta: any) {
  if (!herramienta.disponible) return
  herramientaSeleccionada.value = herramienta
  personaSeleccionada.value = null
  busquedaPersonal.value = ''
  cargarPersonal()
  modalAbierto.value = true
}

function resaltarTexto(texto: string, consulta: string) {
  if (!consulta) return texto
  
  // Normalizamos solo para buscar posiciones, pero mantendremos el texto original
  const textoNorm = normalizar(texto)
  const consultaNorm = normalizar(consulta)
  
  let resultado = ""
  let posActual = 0
  let indice = textoNorm.indexOf(consultaNorm)
  
  if (indice === -1) return texto

  while (indice !== -1) {
    // Añadir lo que hay antes de la coincidencia
    resultado += texto.substring(posActual, indice)
    // Añadir la coincidencia envuelta en <mark> con los caracteres originales
    resultado += `<mark class="bg-green-100 text-green-900 px-0.5 rounded-sm border-b-2 border-green-300 font-bold">${texto.substring(indice, indice + consulta.length)}</mark>`
    
    posActual = indice + consulta.length
    indice = textoNorm.indexOf(consultaNorm, posActual)
  }
  
  resultado += texto.substring(posActual)
  return resultado
}

async function iniciarDevolucion(herramienta: any) {
  if (herramienta.prestadosA.length === 0) return
  herramientaSeleccionada.value = herramienta
  modalDevolucionAbierto.value = true
  
  // Recargar datos frescos al abrir para evitar fechas antiguas
  await recargarDatosHerramientas()
}

async function recargarDatosHerramientas() {
  const toolIdBefore = herramientaSeleccionada.value?.id
  try {
    // Reutilizamos la función de carga principal que ya tiene los endpoints correctos
    await cargarItemsPrestamo()
    
    // Si había una herramienta abierta, recuperamos su nueva versión del array actualizado
    if (toolIdBefore) {
      const actualizada = herramientas.value.find((h: any) => h.id === toolIdBefore)
      if (actualizada) {
        herramientaSeleccionada.value = actualizada
      }
    }
  } catch (error) {
    console.error('Error al sincronizar:', error)
  }
}

async function confirmarPrestamo() {
  if (!personaSeleccionada.value || !herramientaSeleccionada.value) return

  const toolId = herramientaSeleccionada.value.id
  modalAbierto.value = false

  // Activar loader en esta card específica
  procesandoID.value = toolId
  mensajeExito.value = ''
  mensajeError.value = ''

  try {
    const token = obtenerTokenSeguro()
    const respuesta = await fetch(`${API_URL}/prestamos`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      },
      body: JSON.stringify({
        referencia: herramientaSeleccionada.value.codigo,
        descripcion: herramientaSeleccionada.value.nombre,
        ext1: herramientaSeleccionada.value.marca,
        um: herramientaSeleccionada.value.um,
        cedula_operario: personaSeleccionada.value.id,
        nombre_operario: personaSeleccionada.value.nombre,
        cantidad_prestada: 1
      })
    })

    if (!respuesta.ok) {
      const error = await respuesta.json().catch(() => ({}))
      throw new Error(error.error ?? 'Error al crear préstamo')
    }

    const data = await respuesta.json()

    // Actualizar UI local
    herramientaSeleccionada.value.stock -= 1
    herramientaSeleccionada.value.prestadosA.push({
      id_prestamo: data.prestamo.id_prestamo,
      id: personaSeleccionada.value.id,
      nombre: personaSeleccionada.value.nombre,
      cantidad: 1,
      fecha_prestamo: data.prestamo.fecha_prestamo
    })

    // Actualizar contador del personal
    const p = personal.value.find(person => person.id === personaSeleccionada.value.id)
    if (p) p.prestamosActivos += 1

    if (herramientaSeleccionada.value.stock <= 0) {
      herramientaSeleccionada.value.disponible = false
    }

    mensajeExito.value = `Asignado exitosamente: ${herramientaSeleccionada.value.nombre} → ${personaSeleccionada.value.nombre}`
    setTimeout(() => mensajeExito.value = '', 3000)
  } catch (error) {
    mensajeError.value = error instanceof Error ? error.message : 'Error al crear préstamo'
    setTimeout(() => mensajeError.value = '', 5000)
  } finally {
    procesandoID.value = null
  }
}

async function confirmarDevolucion() {
  if (!herramientaSeleccionada.value || !responsableADevolver.value) return

  const responsable = responsableADevolver.value
  const toolId = herramientaSeleccionada.value.id

  // Cerrar formulario y modal si corresponde
  mostrarFormularioDevolucion.value = false

  // Cerrar modal de inmediato si es el último
  if (herramientaSeleccionada.value.prestadosA.length <= 1) {
    modalDevolucionAbierto.value = false
  }

  procesandoID.value = toolId
  mensajeExito.value = ''
  mensajeError.value = ''

  try {
    // Buscar el id_prestamo del préstamo activo de este responsable
    const prestamoActivo = herramientaSeleccionada.value.prestadosA.find(
      (p: any) => p.id === responsable.id && !p.fecha_devolucion
    )

    if (!prestamoActivo || !prestamoActivo.id_prestamo) {
      throw new Error('No se encontró préstamo activo para devolver')
    }

    const token = obtenerTokenSeguro()
    const respuesta = await fetch(`${API_URL}/prestamos/${prestamoActivo.id_prestamo}/devolver`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${token}`
      },
      body: JSON.stringify({
        estado: estadoHerramienta.value,
        observaciones: observacionesHerramienta.value
      })
    })

    if (!respuesta.ok) {
      const error = await respuesta.json().catch(() => ({}))
      throw new Error(error.error ?? 'Error al devolver préstamo')
    }

    // Actualizar UI local
    herramientaSeleccionada.value.stock += 1
    herramientaSeleccionada.value.disponible = true

    // Remover de la lista de prestados de la herramienta
    herramientaSeleccionada.value.prestadosA = herramientaSeleccionada.value.prestadosA.filter(
      (item: any) => item.id !== responsable.id
    )

    // Actualizar contador del personal
    const p = personal.value.find(person => person.id === responsable.id)
    if (p) p.prestamosActivos = Math.max(0, p.prestamosActivos - 1)

    mensajeExito.value = `Devuelto exitosamente: ${herramientaSeleccionada.value.nombre} ← ${responsable.nombre}`
    setTimeout(() => mensajeExito.value = '', 3000)
  } catch (error) {
    mensajeError.value = error instanceof Error ? error.message : 'Error al devolver préstamo'
    setTimeout(() => mensajeError.value = '', 5000)
  } finally {
    procesandoID.value = null
  }
}

function logout() {
  authStore.cerrarSesion()
  router.push('/login')
}

const responsableADevolver = ref<any>(null)
const mostrarFormularioDevolucion = ref(false)
const estadoHerramienta = ref('')
const observacionesHerramienta = ref('')
const menuEstadoAbierto = ref(false)

function prepararDevolucion(responsable: any) {
  responsableADevolver.value = responsable
  mostrarFormularioDevolucion.value = true
  estadoHerramienta.value = ''
  observacionesHerramienta.value = ''
}

function calcularTiempoTranscurrido(fechaStr: string) {
  if (!fechaStr || fechaStr.startsWith('0001')) return ''
  
  const fechaPrestamo = new Date(fechaStr)
  if (isNaN(fechaPrestamo.getTime()) || fechaPrestamo.getFullYear() < 2000) return ''
  
  const ahora = new Date()
  const diffMs = ahora.getTime() - fechaPrestamo.getTime()
  
  if (diffMs < 0) return 'Hace un momento'
  
  const segundos = Math.floor(diffMs / 1000)
  const minutos = Math.floor(segundos / 60)
  const horas = Math.floor(minutos / 60)
  const dias = Math.floor(horas / 24)

  const rMinutos = minutos % 60
  const rHoras = horas % 24

  let partes = []
  if (dias > 0) partes.push(`${dias} ${dias === 1 ? 'día' : 'días'}`)
  if (rHoras > 0) partes.push(`${rHoras} ${rHoras === 1 ? 'hora' : 'horas'}`)
  if (rMinutos > 0 || partes.length === 0) partes.push(`${rMinutos} ${rMinutos === 1 ? 'minuto' : 'minutos'}`)

  return `Hace ${partes.join(', ')}`
}
</script>

<template>
  <div class="min-h-screen bg-white font-sans text-slate-800">

    <!-- ── Navbar ── -->
    <header class="sticky top-0 z-30 w-full flex items-center justify-between px-7 py-4 bg-white/90 border-b border-slate-200 shadow-sm" style="backdrop-filter: blur(14px);">
      <div class="flex items-center gap-3">
        <img src="/favicon.ico" alt="SAO6" class="shrink-0" width="32" height="32" />
        <div class="flex flex-col leading-none gap-0.5">
          <span class="text-[0.82rem] font-extrabold text-slate-800 tracking-widest uppercase">INVENTARIO</span>
          <span class="text-[0.6rem] font-semibold text-slate-500 tracking-widest uppercase">MÓDULO DE PRÉSTAMOS</span>
        </div>
      </div>

      <div class="flex items-center gap-6">
        <!-- Navegación súper básica entre dashboards -->
        <nav v-if="!esOperario" class="hidden md:flex items-center gap-1 bg-slate-100/80 p-1 rounded-full border border-slate-200">
          <router-link to="/dashboard/stock" class="px-4 py-1.5 text-xs font-bold text-slate-500 rounded-full transition-all hover:text-slate-800">
            Stock
          </router-link>
          <router-link to="/dashboard/prestamo" class="px-4 py-1.5 text-xs font-bold bg-white text-green-700 rounded-full shadow-sm border border-slate-200 pointer-events-none">
            Préstamo
          </router-link>
          <router-link to="/dashboard/admin" class="px-4 py-1.5 text-xs font-bold text-slate-500 rounded-full transition-all hover:text-slate-800">
            Admin
          </router-link>
        </nav>

        <div class="hidden sm:flex flex-col items-end leading-none gap-0.5 pl-4 border-l border-slate-200">
          <span class="text-[0.82rem] font-extrabold text-slate-800 tracking-wide uppercase">{{ usuarioLogueado?.nombre_completo || 'USUARIO' }}</span>
          <span class="text-[0.6rem] text-slate-500 tracking-widest uppercase">{{ usuarioLogueado?.rol || 'SIN ROL' }}</span>
        </div>
        <button
          @click="logout"
          class="flex items-center justify-center w-10 h-10 rounded-full bg-white border border-slate-200 text-slate-500 shadow-sm transition-all duration-200 hover:bg-red-50 hover:text-red-500 hover:border-red-200 cursor-pointer"
        >
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"></path>
            <polyline points="16 17 21 12 16 7"></polyline>
            <line x1="21" y1="12" x2="9" y2="12"></line>
          </svg>
        </button>
      </div>
    </header>

    <!-- ── Toast de Notificaciones ── -->
    <Teleport to="body">
      <Transition
        enter-active-class="transition-all duration-300 ease-out"
        enter-from-class="opacity-0 translate-y-[-20px]"
        enter-to-class="opacity-100 translate-y-0"
        leave-active-class="transition-all duration-200 ease-in"
        leave-from-class="opacity-100 translate-y-0"
        leave-to-class="opacity-0 translate-y-[-20px]"
      >
        <div
          v-if="mensajeExito"
          class="fixed top-24 left-1/2 transform -translate-x-1/2 z-[200] px-6 py-3 bg-green-500 text-white rounded-xl shadow-lg flex items-center gap-3 font-bold text-sm"
        >
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
            <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"></path>
            <polyline points="22 4 12 14.01 9 11.01"></polyline>
          </svg>
          {{ mensajeExito }}
        </div>
      </Transition>

      <Transition
        enter-active-class="transition-all duration-300 ease-out"
        enter-from-class="opacity-0 translate-y-[-20px]"
        enter-to-class="opacity-100 translate-y-0"
        leave-active-class="transition-all duration-200 ease-in"
        leave-from-class="opacity-100 translate-y-0"
        leave-to-class="opacity-0 translate-y-[-20px]"
      >
        <div
          v-if="mensajeError"
          class="fixed top-24 left-1/2 transform -translate-x-1/2 z-[200] px-6 py-3 bg-red-500 text-white rounded-xl shadow-lg flex items-center gap-3 font-bold text-sm"
        >
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
            <circle cx="12" cy="12" r="10"></circle>
            <line x1="12" y1="8" x2="12" y2="12"></line>
            <line x1="12" y1="16" x2="12.01" y2="16"></line>
          </svg>
          {{ mensajeError }}
        </div>
      </Transition>
    </Teleport>

    <!-- ── Contenido principal ── -->
    <main class="max-w-7xl mx-auto px-4 lg:px-8 py-10 flex flex-col items-center">

      <!-- Título principal -->
      <div class="text-center flex flex-col items-center gap-3 mb-12">
        <h1 class="text-5xl font-black tracking-tight leading-tight" style="background: linear-gradient(135deg, #0f172a 0%, #334155 100%); -webkit-background-clip: text; -webkit-text-fill-color: transparent; background-clip: text;">
          Catálogo de Préstamos
        </h1>
        <div class="flex items-center gap-2 text-sm">
          <span class="font-medium text-slate-500">Busca y solicita las herramientas que necesites para tu turno.</span>
        </div>
      </div>

      <!-- Barra de búsqueda grande -->
      <div class="w-full max-w-2xl mb-12 relative group">
        <div class="absolute inset-y-0 left-5 flex items-center pointer-events-none">
          <svg class="text-slate-400 group-focus-within:text-green-500 transition-colors" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
            <circle cx="11" cy="11" r="8"></circle>
            <line x1="21" y1="21" x2="16.65" y2="16.65"></line>
          </svg>
        </div>
        <input 
          v-model="busqueda"
          type="text" 
          placeholder="Buscar herramienta por nombre, código o marca..."
          class="w-full py-4 pl-14 pr-6 bg-white border-2 border-slate-200 rounded-full text-slate-700 font-semibold shadow-sm focus:outline-none focus:border-green-400 focus:ring-4 focus:ring-green-100 transition-all placeholder:text-slate-300 placeholder:font-medium"
        />
        <div class="absolute inset-y-0 right-3 flex items-center">
          <span class="bg-slate-100 text-slate-400 text-[0.65rem] font-bold px-3 py-1 rounded-full uppercase tracking-wider">
            {{ herramientasFiltradas.length }} items
          </span>
        </div>
      </div>

      <!-- Controles de Filtros Avanzados -->
      <div class="w-full max-w-5xl mb-10 flex flex-wrap items-center justify-center gap-4 bg-white p-5 rounded-[2.5rem] border border-slate-100 shadow-xl shadow-slate-200/40 relative z-40">
        
        <!-- Filtro Marca (CUSTOM) -->
        <div class="flex flex-col gap-1.5 px-4 border-r border-slate-100 last:border-none relative">
          <label class="text-[0.62rem] font-black text-slate-400 uppercase tracking-widest pl-1">Fabricante</label>
          <div @click="toggleMenu('marca')" class="bg-slate-50 text-[0.75rem] font-extrabold text-slate-700 px-4 py-2.5 rounded-xl flex items-center justify-between gap-3 cursor-pointer min-w-[140px] hover:bg-slate-100 transition-colors border border-slate-200/50">
            <span class="truncate">{{ filtroMarca }}</span>
            <svg class="transition-transform duration-300" :class="{'rotate-180': menuAbierto === 'marca'}" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3"><polyline points="6 9 12 15 18 9"></polyline></svg>
          </div>
          <!-- Menu flotante -->
          <Transition name="fade-slide">
            <div v-if="menuAbierto === 'marca'" class="absolute top-[110%] left-0 w-full min-w-[180px] bg-white border border-slate-100 shadow-2xl rounded-2xl p-2 z-[60]">
              <div 
                v-for="m in marcasUnicas" :key="m"
                @click="filtroMarca = m; menuAbierto = null"
                :class="['px-4 py-2.5 text-xs font-bold rounded-xl cursor-pointer transition-all flex items-center justify-between', filtroMarca === m ? 'bg-green-50 text-green-700' : 'text-slate-600 hover:bg-slate-50']"
              >
                {{ m }}
                <svg v-if="filtroMarca === m" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="4"><polyline points="20 6 9 17 4 12"></polyline></svg>
              </div>
            </div>
          </Transition>
        </div>

        <!-- Filtro Estado -->
        <div class="flex flex-col gap-1.5 px-4 border-r border-slate-100 last:border-none">
          <label class="text-[0.62rem] font-black text-slate-400 uppercase tracking-widest pl-1">Visibilidad</label>
          <div class="flex gap-1 bg-slate-100/50 p-1.5 rounded-xl border border-slate-200/50">
            <button 
              @click="filtroEstado = 'Todos'"
              :class="['px-4 py-1.5 text-[0.6rem] font-black rounded-lg transition-all', filtroEstado === 'Todos' ? 'bg-white text-slate-800 shadow-sm border border-slate-100' : 'text-slate-400 hover:text-slate-600']"
            >TODOS</button>
            <button 
              @click="filtroEstado = 'Disponibles'"
              :class="['px-4 py-1.5 text-[0.6rem] font-black rounded-lg transition-all', filtroEstado === 'Disponibles' ? 'bg-[#22c55e] text-white shadow-md' : 'text-slate-400 hover:text-slate-600']"
            >LIBRES</button>
            <button 
              @click="filtroEstado = 'Prestados'"
              :class="['px-4 py-1.5 text-[0.6rem] font-black rounded-lg transition-all', filtroEstado === 'Prestados' ? 'bg-emerald-600 text-white shadow-md' : 'text-slate-400 hover:text-slate-600']"
            >EN USO</button>
          </div>
        </div>

        <!-- Ordenamiento Stock (CUSTOM) -->
        <div class="flex flex-col gap-1.5 px-4 relative">
          <label class="text-[0.62rem] font-black text-slate-400 uppercase tracking-widest pl-1">Ordenar por</label>
          <div @click="toggleMenu('orden')" class="bg-slate-50 text-[0.75rem] font-extrabold text-slate-700 px-4 py-2.5 rounded-xl flex items-center justify-between gap-3 cursor-pointer min-w-[160px] hover:bg-slate-100 transition-colors border border-slate-200/50">
            <span class="truncate">{{ ordenStock === 'default' ? 'Nombre (A-Z)' : ordenStock === 'mayor-menor' ? 'Stock: Mayor a Menor' : 'Stock: Menor a Mayor' }}</span>
            <svg class="transition-transform duration-300" :class="{'rotate-180': menuAbierto === 'orden'}" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3"><polyline points="6 9 12 15 18 9"></polyline></svg>
          </div>
          <Transition name="fade-slide">
            <div v-if="menuAbierto === 'orden'" class="absolute top-[110%] right-0 w-full min-w-[200px] bg-white border border-slate-100 shadow-2xl rounded-2xl p-2 z-[60]">
              <div 
                @click="ordenStock = 'default'; menuAbierto = null"
                :class="['px-4 py-2.5 text-xs font-bold rounded-xl cursor-pointer transition-all flex items-center justify-between', ordenStock === 'default' ? 'bg-green-50 text-green-700' : 'text-slate-600 hover:bg-slate-50']"
              >
                Nombre (A-Z)
                <svg v-if="ordenStock === 'default'" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="4"><polyline points="20 6 9 17 4 12"></polyline></svg>
              </div>
              <div 
                @click="ordenStock = 'mayor-menor'; menuAbierto = null"
                :class="['px-4 py-2.5 text-xs font-bold rounded-xl cursor-pointer transition-all flex items-center justify-between', ordenStock === 'mayor-menor' ? 'bg-green-50 text-green-700' : 'text-slate-600 hover:bg-slate-50']"
              >
                Stock: Mayor a Menor
                <svg v-if="ordenStock === 'mayor-menor'" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="4"><polyline points="20 6 9 17 4 12"></polyline></svg>
              </div>
              <div 
                @click="ordenStock = 'menor-mayor'; menuAbierto = null"
                :class="['px-4 py-2.5 text-xs font-bold rounded-xl cursor-pointer transition-all flex items-center justify-between', ordenStock === 'menor-mayor' ? 'bg-green-50 text-green-700' : 'text-slate-600 hover:bg-slate-50']"
              >
                Stock: Menor a Mayor
                <svg v-if="ordenStock === 'menor-mayor'" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="4"><polyline points="20 6 9 17 4 12"></polyline></svg>
              </div>
            </div>
          </Transition>
        </div>

        <!-- Botón Reset -->
        <button 
          @click="filtroMarca = 'Todas'; filtroEstado = 'Todos'; ordenStock = 'default'; busqueda = ''; menuAbierto = null"
          class="ml-2 w-11 h-11 rounded-2xl bg-white border border-slate-100 flex items-center justify-center text-slate-400 hover:bg-red-50 hover:text-red-500 hover:border-red-100 transition-all shadow-sm group"
          title="Limpiar Filtros"
        >
          <svg class="group-hover:rotate-90 transition-transform duration-300" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round"><path d="M18 6L6 18M6 6l12 12"></path></svg>
        </button>
      </div>

      <!-- Grid de Skeletons Nativos (Solo mientras carga) -->
      <div v-if="cargando" class="w-full grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
        <div 
          v-for="i in 8" 
          :key="i"
          class="bg-white rounded-3xl p-6 border border-slate-100 shadow-sm flex flex-col items-start relative overflow-hidden"
        >
          <!-- Header skeleton -->
          <div class="w-full flex justify-between items-start mb-5 relative z-10 animate-pulse">
            <div class="w-12 h-12 rounded-2xl bg-slate-100 shrink-0"></div>
            <div class="flex flex-col items-end gap-2 mt-1">
              <div class="w-20 h-4 bg-slate-100 rounded-full"></div>
              <div class="w-14 h-3 bg-slate-50 rounded-full"></div>
            </div>
          </div>
          <!-- Info principal skeleton -->
          <div class="mb-6 z-10 flex-1 w-full animate-pulse mt-1">
            <div class="w-4/5 h-6 bg-slate-200/60 rounded-lg mb-2"></div>
            <div class="w-2/5 h-3 bg-slate-100 rounded-full"></div>
          </div>
          <!-- Botón skeleton -->
          <div class="w-full h-11 rounded-xl bg-slate-100 animate-pulse z-10 mt-auto"></div>
        </div>
      </div>

      <!-- Grid de Herramientas Reales -->
      <TransitionGroup 
        v-else-if="herramientasFiltradas.length > 0"
        name="staggered-fade" 
        tag="div" 
        class="w-full grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6"
      >
          <div 
            v-for="(herramienta, index) in herramientasFiltradas" 
            :key="herramienta.id"
            class="bg-white rounded-3xl p-6 border border-slate-200 shadow-sm hover:shadow-xl transition-all duration-300 group flex flex-col items-start relative overflow-hidden"
            :style="{ transitionDelay: `${index * 40}ms` }"
          >
            <!-- Efecto hover de fondo -->
            <div class="absolute top-0 right-0 p-32 bg-gradient-to-bl from-green-50/50 to-transparent rounded-full -mr-16 -mt-16 opacity-0 group-hover:opacity-100 transition-opacity duration-500 pointer-events-none"></div>

            <!-- Loader Overlay para la card -->
            <div 
              v-if="procesandoID === herramienta.id"
              class="absolute inset-0 z-50 bg-white/60 backdrop-blur-[2px] flex items-center justify-center flex-col gap-3 transition-opacity duration-300"
            >
              <div class="w-10 h-10 border-4 border-slate-100 border-t-green-500 rounded-full animate-spin"></div>
              <span class="text-[0.6rem] font-black text-slate-500 uppercase tracking-widest animate-pulse">Procesando...</span>
            </div>

            <!-- Header de tarjeta -->
            <div class="w-full flex justify-between items-start mb-5 relative z-10">
              <div class="w-12 h-12 rounded-2xl flex items-center justify-center shrink-0 border transition-colors"
                :class="herramienta.disponible ? 'bg-green-50 border-green-100' : 'bg-slate-50 border-slate-100'">
                <svg v-if="herramienta.disponible" width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="#22c55e" stroke-width="2"><path d="M14.7 6.3a1 1 0 0 0 0 1.4l1.6 1.6a1 1 0 0 0 1.4 0l3.77-3.77a6 6 0 0 1-7.94 9.36l-7.1 7.1a1 1 0 0 1-1.4 0l-2.83-2.83a1 1 0 0 1 0-1.4l7.1-7.1a6 6 0 0 1 9.36-7.94l-3.76 3.76z"></path></svg>
                <svg v-else width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="#94a3b8" stroke-width="2"><rect x="3" y="11" width="18" height="11" rx="2" ry="2"></rect><path d="M7 11V7a5 5 0 0 1 10 0v4"></path></svg>
              </div>
              <div class="flex flex-col items-end">
                <span :class="[
                  'text-[0.6rem] font-extrabold px-2.5 py-1 rounded-full uppercase tracking-wider',
                  herramienta.disponible ? 'bg-green-100 text-green-700' : 'bg-red-50 text-red-500'
                ]">
                  {{ herramienta.disponible ? 'DISPONIBLE' : 'NO DISPONIBLE' }}
                </span>
                <span class="text-xs font-bold text-slate-400 mt-1.5" v-if="herramienta.disponible">
                  Stock: <span class="text-slate-600">{{ herramienta.stock }} {{ herramienta.um }}</span>
                </span>
              </div>
            </div>

            <!-- Info principal -->
            <div class="mb-6 z-10 flex-1">
              <h3 class="text-lg font-black text-slate-800 leading-tight mb-1 group-hover:text-green-600 transition-colors" v-html="resaltarTexto(herramienta.nombre, busqueda)"></h3>
              <p class="text-sm font-semibold text-slate-400" v-html="resaltarTexto(herramienta.codigo, busqueda)"></p>
            </div>

            <div class="w-full flex items-center justify-between mb-6 z-10">
              <div class="flex flex-col">
                <span class="text-[0.6rem] font-bold text-slate-400 uppercase tracking-wider mb-0.5">MARCA</span>
                <span class="text-sm font-extrabold text-slate-700" v-html="resaltarTexto(herramienta.marca, busqueda)"></span>
              </div>
            </div>

            <!-- Botones de acción -->
            <div class="w-full flex gap-3 mt-auto relative z-10 pt-2">
              <button
                @click="solicitarPrestamo(herramienta)"
                :disabled="!herramienta.disponible"
                :class="[
                  'flex-1 py-3.5 rounded-2xl font-black text-[0.65rem] uppercase tracking-widest transition-all duration-300 flex items-center justify-center gap-2',
                  herramienta.disponible 
                    ? 'bg-[#22c55e] text-white hover:brightness-95 hover:shadow-[0_8px_16px_rgba(34,197,94,0.3)] hover:-translate-y-0.5 active:scale-95 cursor-pointer' 
                    : 'bg-slate-100 text-slate-300 cursor-not-allowed opacity-70'
                ]"
              >
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3.5" stroke-linecap="round" stroke-linejoin="round"><line x1="12" y1="5" x2="12" y2="19"></line><line x1="5" y1="12" x2="19" y2="12"></line></svg>
                Prestar
              </button>
              
              <button
                @click="iniciarDevolucion(herramienta)"
                :disabled="herramienta.prestadosA.length === 0"
                :class="[
                  'flex-1 py-3.5 rounded-2xl font-black text-[0.65rem] uppercase tracking-widest transition-all duration-300 flex items-center justify-center gap-2',
                  herramienta.prestadosA.length > 0
                    ? 'bg-emerald-600 text-white hover:brightness-95 hover:shadow-[0_8px_16px_rgba(5,150,105,0.3)] hover:-translate-y-0.5 active:scale-95 cursor-pointer'
                    : 'bg-slate-100 text-slate-300 cursor-not-allowed opacity-70'
                ]"
              >
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3.5" stroke-linecap="round" stroke-linejoin="round"><path d="M2.5 2v6h6M2.5 8a10 10 0 1 1 2.36 5.06"></path></svg>
                Retorno
              </button>
            </div>
          </div>
      </TransitionGroup>
      
      <!-- Error state -->
      <div v-else-if="errorCarga" class="flex flex-col items-center justify-center p-12 text-center">
        <div class="w-20 h-20 bg-red-50 rounded-full flex items-center justify-center mb-4">
          <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="#ef4444" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"></circle><line x1="12" y1="8" x2="12" y2="12"></line><line x1="12" y1="16" x2="12.01" y2="16"></line></svg>
        </div>
        <h3 class="text-xl font-bold text-red-700 mb-2">Error al cargar datos</h3>
        <p class="text-sm font-medium text-slate-500">{{ errorCarga }}</p>
        <button @click="cargarItemsPrestamo" class="mt-4 px-6 py-2 bg-green-500 text-white rounded-xl text-xs font-black uppercase tracking-widest hover:bg-green-600 transition-all">
          Reintentar
        </button>
      </div>

      <!-- Empty state -->
      <div v-else class="flex flex-col items-center justify-center p-12 text-center">
        <div class="w-20 h-20 bg-slate-100 rounded-full flex items-center justify-center mb-4">
          <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="#94a3b8" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="11" cy="11" r="8"></circle><line x1="21" y1="21" x2="16.65" y2="16.65"></line></svg>
        </div>
        <h3 class="text-xl font-bold text-slate-700 mb-2">No se encontraron herramientas</h3>
        <p class="text-sm font-medium text-slate-500">Intenta buscar con otros términos o verifica el código.</p>
      </div>

    </main>

    <!-- Modal de Asignación de Préstamo via Teleport -->
    <Teleport to="body">
      <Transition
        enter-active-class="transition-opacity duration-300 ease-out"
        enter-from-class="opacity-0"
        enter-to-class="opacity-100"
        leave-active-class="transition-opacity duration-200 ease-in"
        leave-from-class="opacity-100"
        leave-to-class="opacity-0"
      >
        <div
          v-if="modalAbierto"
          class="fixed inset-0 z-[100] flex items-center justify-center px-4"
        >
          <!-- Fondo oscuro responsivo (Backdrop) -->
          <div 
            class="absolute inset-0" 
            style="background: rgba(15, 23, 42, 0.45); backdrop-filter: blur(2px);"
            @click="modalAbierto = false"
          ></div>

          <!-- Contenedor del Modal (Animado con Tailwind efecto resorte) -->
          <Transition
            appear
            enter-active-class="transition-all duration-[450ms] ease-[cubic-bezier(0.34,1.56,0.64,1)] transform"
            enter-from-class="opacity-0 scale-75 translate-y-12"
            enter-to-class="opacity-100 scale-100 translate-y-0"
            leave-active-class="transition-all duration-200 ease-in transform"
            leave-from-class="opacity-100 scale-100 translate-y-0"
            leave-to-class="opacity-0 scale-95 translate-y-4"
          >
            <div
              v-if="modalAbierto"
              class="relative bg-white rounded-3xl shadow-2xl w-full max-w-3xl overflow-hidden z-10 flex flex-col max-h-[90vh]"
              @click.stop
            >
              <!-- Header del modal (Minimalista) -->
              <div class="flex items-center justify-between px-8 pt-8 pb-6 shrink-0 relative z-10">
                <div class="flex items-center gap-5 min-w-0">
                  <div class="w-14 h-14 rounded-full bg-green-50 flex items-center justify-center shrink-0">
                    <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="#22c55e" stroke-width="2"><path d="M16 21V5a2 2 0 0 0-2-2h-4a2 2 0 0 0-2 2v16"></path></svg>
                  </div>
                  <div class="flex-1 min-w-0">
                    <h2 class="text-2xl font-black text-slate-800 leading-tight">Asignar Préstamo</h2>
                    <p class="text-sm text-slate-500 font-medium mt-1 truncate flex items-center gap-2">
                      <span>{{ herramientaSeleccionada?.nombre }}</span> 
                      <span class="w-1.5 h-1.5 rounded-full bg-slate-300"></span> 
                      <span class="font-bold text-slate-400">{{ herramientaSeleccionada?.codigo }}</span>
                    </p>
                  </div>
                </div>
                <button
                  @click="modalAbierto = false"
                  class="shrink-0 flex items-center justify-center w-10 h-10 rounded-full text-slate-400 hover:bg-red-50 hover:text-red-500 transition-all duration-200 cursor-pointer border border-transparent hover:border-red-100"
                >
                  <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round"><line x1="18" y1="6" x2="6" y2="18"></line><line x1="6" y1="6" x2="18" y2="18"></line></svg>
                </button>
              </div>

              <!-- Searchbar Minimalista -->
              <div class="px-8 pb-6 shrink-0">
                <div class="relative group">
                  <div class="absolute inset-y-0 left-0 flex items-center pl-4 pointer-events-none">
                    <svg class="text-slate-400 group-focus-within:text-green-500 transition-colors" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><circle cx="11" cy="11" r="8"></circle><line x1="21" y1="21" x2="16.65" y2="16.65"></line></svg>
                  </div>
                  <input 
                    v-model="busquedaPersonal"
                    type="text" 
                    placeholder="Busca por nombre o cédula..."
                    class="w-full py-3.5 pl-12 pr-4 bg-transparent border-b-2 border-slate-100 text-[1.1rem] text-slate-700 font-semibold focus:outline-none focus:border-green-400 transition-all placeholder:text-slate-300"
                  />
                </div>
              </div>

              <!-- Lista de Personal Simple -->
              <div class="overflow-y-auto flex-1 px-8 pb-6">
                <!-- Loader -->
                <div v-if="cargandoPersonal" class="py-12 flex flex-col items-center">
                  <div class="w-12 h-12 border-4 border-slate-100 border-t-green-500 rounded-full animate-spin mb-4"></div>
                  <p class="text-sm font-bold text-slate-400 uppercase tracking-widest">Cargando personal...</p>
                </div>

                <!-- Error -->
                <div v-else-if="errorPersonal" class="py-12 flex flex-col items-center">
                  <div class="w-16 h-16 bg-red-50 rounded-full flex items-center justify-center mb-3">
                    <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="#ef4444" stroke-width="2"><circle cx="12" cy="12" r="10"></circle><line x1="12" y1="8" x2="12" y2="12"></line><line x1="12" y1="16" x2="12.01" y2="16"></line></svg>
                  </div>
                  <p class="text-base font-bold text-red-600 mb-2">Error al cargar personal</p>
                  <p class="text-sm text-slate-500 mb-4">{{ errorPersonal }}</p>
                  <button @click="cargarPersonal" class="px-4 py-2 bg-green-500 text-white rounded-xl text-xs font-black uppercase tracking-widest hover:bg-green-600 transition-all">
                    Reintentar
                  </button>
                </div>

                <ul v-else class="flex flex-col gap-3">
                  <li
                    v-for="p in personalFiltrado"
                    :key="p.id"
                    @click="personaSeleccionada = p"
                    :class="[
                      'flex items-center gap-4 p-4 rounded-2xl cursor-pointer transition-all duration-200 group border',
                      personaSeleccionada?.id === p.id
                        ? 'border-green-400 bg-green-50/30 shadow-[0_4px_20px_rgba(34,197,94,0.08)]'
                        : 'border-slate-100 bg-white hover:border-slate-200 hover:shadow-sm'
                    ]"
                  >
                    <!-- Check radio minimalista -->
                    <span class="shrink-0 flex items-center justify-center w-6 h-6">
                      <span
                        v-if="personaSeleccionada?.id === p.id"
                        class="w-5 h-5 bg-green-500 rounded-full flex items-center justify-center text-white"
                      >
                        <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"><polyline points="20 6 9 17 4 12"></polyline></svg>
                      </span>
                      <span v-else class="w-5 h-5 rounded-full border-2 border-slate-200 group-hover:border-green-300"></span>
                    </span>

                    <!-- Avatar/Foto -->
                    <img
                      :src="p.foto"
                      @error="manejarErrorFoto($event, p)"
                      alt="Avatar"
                      class="w-12 h-12 rounded-full border border-slate-200 shadow-sm shrink-0 bg-slate-50 object-cover"
                    />

                    <!-- Información Principal -->
                    <div class="flex-1 min-w-0 flex flex-col justify-center">
                      <p :class="[
                        'text-[0.95rem] font-extrabold leading-tight truncate transition-colors duration-200', 
                        personaSeleccionada?.id === p.id ? 'text-green-800' : 'text-slate-700 group-hover:text-slate-900'
                      ]">
                        {{ p.nombre }}
                      </p>
                      <div class="flex items-center gap-2 mt-1 truncate">
                        <p class="text-[0.75rem] font-bold text-slate-400">CC: {{ p.id }}</p>
                        <span class="w-1 h-1 rounded-full bg-slate-300"></span>
                        <span :class="[ 
                          'text-[0.6rem] font-black rounded-md px-1.5 py-0.5 tracking-wider uppercase',
                          p.rol === 'OPERARIO' ? 'bg-blue-50 text-blue-600' : p.rol === 'TÉCNICO' ? 'bg-purple-50 text-purple-600' : 'bg-orange-50 text-orange-600'
                        ]">
                          {{ p.rol }}
                        </span>
                      </div>
                    </div>

                    <!-- Indicador Préstamos -->
                    <div class="flex flex-col items-end justify-center shrink-0">
                      <div :class="[
                        'flex items-center gap-1.5 px-3 py-1.5 rounded-lg border shadow-sm transition-colors',
                        p.prestamosActivos > 0 
                          ? 'bg-emerald-50 border-emerald-200 text-emerald-700 shadow-emerald-500/10' 
                          : 'bg-slate-50 border-slate-200 text-slate-500'
                      ]">
                        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"></path><polyline points="7 10 12 15 17 10"></polyline><line x1="12" y1="15" x2="12" y2="3"></line></svg>
                        <span class="text-sm font-black">{{ p.prestamosActivos }}</span>
                      </div>
                      <span class="text-[0.55rem] font-bold text-slate-400 uppercase mt-1 tracking-widest">{{ p.prestamosActivos === 1 ? 'Prestado' : 'Prestados' }}</span>
                    </div>

                  </li>
                </ul>
                
                <div v-if="personalFiltrado.length === 0" class="py-12 flex flex-col items-center">
                   <div class="w-16 h-16 bg-slate-200 rounded-full flex items-center justify-center mb-3">
                     <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="#64748b" stroke-width="2"><circle cx="11" cy="11" r="8"></circle><line x1="21" y1="21" x2="16.65" y2="16.65"></line></svg>
                   </div>
                   <p class="text-base font-bold text-slate-500">No se encontraron empleados coincidentes.</p>
                </div>
              </div>

              <!-- Footer de acción -->
              <div class="px-8 py-6 border-t border-slate-100 shrink-0 bg-white relative">
                <button
                  @click="confirmarPrestamo"
                  :disabled="!personaSeleccionada"
                  :class="[
                    'w-full py-4 rounded-xl font-black text-base tracking-wide transition-all duration-300 flex items-center justify-center gap-2 relative z-10',
                    personaSeleccionada
                      ? 'bg-[#22c55e] border-2 border-[#22c55e] text-white hover:brightness-95 hover:shadow-[0_8px_20px_rgba(34,197,94,0.4)] hover:-translate-y-0.5 cursor-pointer'
                      : 'bg-slate-50 text-slate-400 cursor-not-allowed border-2 border-slate-200 border-dashed hover:bg-slate-100'
                  ]"
                >
                  <svg v-if="personaSeleccionada" width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"></path><polyline points="22 4 12 14.01 9 11.01"></polyline></svg>
                  {{ personaSeleccionada ? `Confirmar Asignación para ${personaSeleccionada.nombre.split(' ')[0]}` : 'Seleccione a un Responsable para autorizar' }}
                </button>
              </div>
            </div>
          </Transition>
        </div>
      </Transition>
    </Teleport>

    <!-- Modal de Devolución via Teleport -->
    <Teleport to="body">
      <Transition
        enter-active-class="transition-opacity duration-300 ease-out"
        enter-from-class="opacity-0"
        enter-to-class="opacity-100"
        leave-active-class="transition-opacity duration-200 ease-in"
        leave-from-class="opacity-100"
        leave-to-class="opacity-0"
      >
        <div
          v-if="modalDevolucionAbierto"
          class="fixed inset-0 z-[100] flex items-center justify-center px-4"
        >
          <div 
            class="absolute inset-0" 
            style="background: rgba(15, 23, 42, 0.45); backdrop-filter: blur(2px);"
            @click="modalDevolucionAbierto = false"
          ></div>

          <Transition
            appear
            enter-active-class="transition-all duration-[450ms] ease-[cubic-bezier(0.34,1.56,0.64,1)] transform"
            enter-from-class="opacity-0 scale-75 translate-y-12"
            enter-to-class="opacity-100 scale-100 translate-y-0"
            leave-active-class="transition-all duration-200 ease-in transform"
            leave-from-class="opacity-100 scale-100 translate-y-0"
            leave-to-class="opacity-0 scale-95 translate-y-4"
          >
            <div
              v-if="modalDevolucionAbierto"
              class="relative bg-white rounded-3xl shadow-2xl w-full max-w-2xl overflow-hidden z-10 flex flex-col max-h-[80vh]"
              @click.stop
            >
              <!-- Header -->
              <div class="px-8 pt-8 pb-6 border-b border-slate-50 shrink-0 flex items-center justify-between">
                <div class="flex items-center gap-4 text-green-600">
                  <div class="w-12 h-12 bg-green-50 rounded-2xl flex items-center justify-center">
                    <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><path d="M12 2v10l4.5 4.5"></path><circle cx="12" cy="12" r="10"></circle></svg>
                  </div>
                  <div>
                    <h2 class="text-xl font-black text-slate-800">Gestionar Devolución</h2>
                    <p class="text-xs font-bold text-slate-400 mt-0.5">{{ herramientaSeleccionada?.nombre }}</p>
                  </div>
                </div>
                
                <div class="flex items-center gap-2">
                  <button 
                    @click="recargarDatosHerramientas"
                    class="p-2 rounded-full hover:bg-slate-100 text-slate-400 transition-all active:scale-90 group"
                    title="Sincronizar datos"
                  >
                    <svg class="group-hover:rotate-180 transition-transform duration-500" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="M23 4v6h-6"></path><path d="M1 20v-6h6"></path><path d="M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15"></path></svg>
                  </button>
                  <button @click="modalDevolucionAbierto = false" class="p-2 rounded-full hover:bg-slate-100 text-slate-400 transition-colors">
                    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><line x1="18" y1="6" x2="6" y2="18"></line><line x1="6" y1="6" x2="18" y2="18"></line></svg>
                  </button>
                </div>
              </div>

              <!-- Lista de personas o Formulario -->
              <div class="p-6 overflow-y-auto min-h-[300px]">
                
                <!-- Vista de Lista de Responsables -->
                <div v-if="!mostrarFormularioDevolucion">
                  <div class="flex items-center gap-3 mb-6">
                    <div class="w-10 h-10 rounded-xl bg-gradient-to-br from-green-400 to-green-600 flex items-center justify-center shadow-lg shadow-green-500/20">
                      <svg class="w-5 h-5 text-white" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0z"/>
                      </svg>
                    </div>
                    <div>
                      <p class="text-sm font-black text-slate-800">Responsables Actuales</p>
                      <p class="text-xs font-medium text-slate-400">{{ herramientaSeleccionada?.prestadosA?.length || 0 }} persona(s) con esta herramienta</p>
                    </div>
                  </div>

                  <div class="space-y-3">
                    <div
                      v-for="(responsable, index) in herramientaSeleccionada?.prestadosA"
                      :key="responsable.id"
                      class="group relative overflow-hidden rounded-2xl border border-slate-100 bg-white hover:border-green-200 hover:shadow-xl hover:shadow-green-500/10 transition-all duration-300"
                      :style="{ animationDelay: `${(index as number) * 100}ms` }"
                    >
                      <!-- Fondo decorativo -->
                      <div class="absolute inset-0 bg-gradient-to-r from-green-50/0 via-green-50/0 to-green-50/30 opacity-0 group-hover:opacity-100 transition-opacity duration-500"></div>

                      <div class="relative p-4 flex items-center gap-4">
                        <div class="relative shrink-0">
                          <img
                            :src="construirUrlFoto(responsable.id)"
                            :alt="responsable.nombre"
                            class="w-12 h-12 rounded-full object-cover border-2 border-green-200 bg-green-50"
                            @error="manejarErrorFoto($event, responsable)"
                          />
                        </div>

                        <div class="flex-1 min-w-0">
                          <p class="text-sm font-bold text-slate-800">{{ responsable.nombre }}</p>
                          <p class="text-xs text-slate-500 mt-0.5">CC: {{ responsable.id }}</p>
                          <div v-if="responsable.fecha_prestamo && !responsable.fecha_prestamo.startsWith('0001')" class="flex flex-col mt-0.5">
                            <p class="text-[0.7rem] text-green-600 font-bold uppercase tracking-wide">
                              Desde {{ new Date(responsable.fecha_prestamo).toLocaleDateString('es-CO', { day: 'numeric', month: 'short' }) }} - {{ new Date(responsable.fecha_prestamo).toLocaleTimeString('es-CO', { hour: '2-digit', minute: '2-digit' }) }}
                            </p>
                            <p class="text-[0.7rem] text-slate-400 font-extrabold italic">
                              {{ calcularTiempoTranscurrido(responsable.fecha_prestamo) }}
                            </p>
                          </div>
                          <div v-else class="mt-1">
                            <span class="text-[0.6rem] font-bold text-slate-300 animate-pulse uppercase">Cargando fecha...</span>
                          </div>
                        </div>

                        <button
                          @click="prepararDevolucion(responsable)"
                          :disabled="procesandoID === herramientaSeleccionada?.id"
                          class="shrink-0 px-4 py-2.5 rounded-xl bg-slate-800 text-white text-xs font-bold shadow-md hover:bg-green-500 hover:shadow-green-500/30 hover:-translate-y-0.5 disabled:opacity-50 disabled:cursor-not-allowed disabled:hover:translate-y-0 transition-all duration-200 flex items-center gap-2"
                        >
                          <svg class="w-4 h-4" fill="none" stroke="currentColor" stroke-width="2.5" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"/>
                          </svg>
                          Devolver
                        </button>
                      </div>
                    </div>
                  </div>
                </div>

                <!-- Formulario de Detalles de Devolución -->
                <div v-else class="flex flex-col gap-6 animate-in fade-in slide-in-from-bottom-4 duration-300">
                  <div class="flex items-center gap-4 p-4 bg-slate-50 rounded-2xl border border-slate-100">
                    <img :src="construirUrlFoto(responsableADevolver.id)" class="w-16 h-16 rounded-full border-2 border-green-200 object-cover shadow-sm bg-white" @error="manejarErrorFoto($event, responsableADevolver)"/>
                    <div class="flex-1 min-w-0">
                      <p class="text-[0.95rem] font-black text-slate-800 truncate">Devolución de {{ responsableADevolver.nombre }}</p>
                      <div class="flex flex-col gap-0.5 mt-1">
                        <div v-if="responsableADevolver.fecha_prestamo && !responsableADevolver.fecha_prestamo.startsWith('0001')" class="flex flex-col gap-0.5">
                          <div class="flex items-center gap-1.5 text-[0.68rem] font-bold text-green-600 uppercase tracking-wide">
                            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3"><rect x="3" y="4" width="18" height="18" rx="2" ry="2"></rect><line x1="16" y1="2" x2="16" y2="6"></line><line x1="8" y1="2" x2="8" y2="6"></line><line x1="3" y1="10" x2="21" y2="10"></line></svg>
                            <span>Préstamo: {{ new Date(responsableADevolver.fecha_prestamo).toLocaleDateString('es-CO', { day: 'numeric', month: 'short' }) }} - {{ new Date(responsableADevolver.fecha_prestamo).toLocaleTimeString('es-CO', { hour: '2-digit', minute: '2-digit' }) }}</span>
                          </div>
                          <div class="flex items-center gap-1.5 text-[0.68rem] font-black text-slate-400 italic">
                            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3"><circle cx="12" cy="12" r="10"></circle><polyline points="12 6 12 12 16 14"></polyline></svg>
                            <span>Tiempo en uso: {{ calcularTiempoTranscurrido(responsableADevolver.fecha_prestamo) }}</span>
                          </div>
                        </div>
                        <div v-else class="py-1">
                          <span class="text-[0.65rem] font-black text-slate-300 animate-pulse uppercase tracking-widest">Sincronizando tiempo de préstamo...</span>
                        </div>
                      </div>
                    </div>
                  </div>

                  <div class="flex flex-col gap-2 relative">
                    <label class="text-[0.65rem] font-black text-slate-400 uppercase tracking-widest pl-1">Estado de la Herramienta (Opcional)</label>
                    <div 
                      @click="menuEstadoAbierto = !menuEstadoAbierto"
                      class="w-full p-4 bg-white border-2 border-slate-100 rounded-2xl text-sm font-bold text-slate-700 flex items-center justify-between cursor-pointer hover:border-green-200 transition-all shadow-sm"
                      :class="{'border-green-400 ring-4 ring-green-50': menuEstadoAbierto}"
                    >
                      <span :class="{'text-slate-300': !estadoHerramienta}">{{ estadoHerramienta || 'Seleccione estado...' }}</span>
                      <svg class="transition-transform duration-300" :class="{'rotate-180': menuEstadoAbierto}" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><polyline points="6 9 12 15 18 9"></polyline></svg>
                    </div>

                    <!-- Opciones personalizadas -->
                    <Transition name="fade-slide">
                      <div v-if="menuEstadoAbierto" class="absolute top-[105%] left-0 w-full bg-white border border-slate-100 shadow-2xl rounded-2xl p-2 z-[70]">
                        <div 
                          v-for="opcion in ['Mal estado', 'Sucia', 'No entrega']" :key="opcion"
                          @click="estadoHerramienta = opcion; menuEstadoAbierto = false"
                          :class="[
                            'px-4 py-3 text-sm font-bold rounded-xl cursor-pointer transition-all flex items-center justify-between',
                            estadoHerramienta === opcion ? 'bg-green-50 text-green-700' : 'text-slate-600 hover:bg-slate-50'
                          ]"
                        >
                          {{ opcion }}
                          <svg v-if="estadoHerramienta === opcion" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="4"><polyline points="20 6 9 17 4 12"></polyline></svg>
                        </div>
                        <div 
                          @click="estadoHerramienta = ''; menuEstadoAbierto = false"
                          class="px-4 py-3 text-sm font-bold text-slate-400 hover:bg-red-50 hover:text-red-500 rounded-xl cursor-pointer transition-all mt-1 border-t border-slate-50"
                        >
                          Limpiar selección
                        </div>
                      </div>
                    </Transition>
                  </div>

                  <div class="flex flex-col gap-2">
                    <label class="text-[0.65rem] font-black text-slate-400 uppercase tracking-widest pl-1">Observaciones / Comentarios</label>
                    <textarea 
                      v-model="observacionesHerramienta"
                      rows="3"
                      placeholder="Escribe aquí cualquier detalle relevante sobre el estado físico o funcional..."
                      class="w-full p-4 bg-white border-2 border-slate-100 rounded-2xl text-sm font-medium text-slate-700 focus:outline-none focus:border-green-400 focus:ring-4 focus:ring-green-50 transition-all resize-none placeholder:text-slate-300"
                    ></textarea>
                  </div>

                  <div class="flex gap-3 pt-2">
                    <button 
                      @click="mostrarFormularioDevolucion = false"
                      class="flex-1 py-4 rounded-2xl bg-slate-100 text-slate-400 text-[0.7rem] font-black uppercase tracking-widest hover:bg-slate-200 hover:text-slate-600 transition-all active:scale-95"
                    >
                      Cancelar
                    </button>
                    <button 
                      @click="confirmarDevolucion"
                      class="flex-[2] py-4 rounded-2xl bg-green-600 text-white text-[0.7rem] font-black uppercase tracking-widest shadow-lg shadow-green-500/20 hover:brightness-95 hover:-translate-y-0.5 transition-all active:scale-95 flex items-center justify-center gap-2"
                    >
                      <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"><path d="M19 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11l5 5v11a2 2 0 0 1-2 2z"></path><polyline points="17 21 17 13 7 13 7 21"></polyline><polyline points="7 3 7 8 15 8"></polyline></svg>
                      Guardar Devolución
                    </button>
                  </div>
                </div>
              </div>

              <div class="p-6 bg-slate-50 border-t border-slate-100 text-center">
                <p class="text-[0.65rem] font-bold text-slate-400 uppercase tracking-widest italic">Asegúrate de verificar el estado físico de la herramienta antes de procesar.</p>
              </div>
            </div>
          </Transition>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<style scoped>
/* Transición en cascada para las tarjetas */
.staggered-fade-enter-active {
  transition: all 0.5s cubic-bezier(0.34, 1.56, 0.64, 1);
}
.staggered-fade-leave-active {
  transition: all 0.4s cubic-bezier(0.4, 0, 0.2, 1);
  position: absolute;
}
.staggered-fade-enter-from,
.staggered-fade-leave-to {
  opacity: 0;
  transform: translateY(30px) scale(0.9);
}
.staggered-fade-move {
  transition: transform 0.5s cubic-bezier(0.34, 1.56, 0.64, 1);
}

/* Animación para los dropdowns custom */
.fade-slide-enter-active {
  transition: all 0.3s cubic-bezier(0.34, 1.56, 0.64, 1);
}
.fade-slide-leave-active {
  transition: all 0.2s ease-in;
}
.fade-slide-enter-from,
.fade-slide-leave-to {
  opacity: 0;
  transform: translateY(-10px) scale(0.95);
}
</style>
