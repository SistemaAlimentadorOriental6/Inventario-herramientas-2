<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, nextTick, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../../auth/store/useAuthStore'
import { stockService } from '../services/stockService'
import type { DetalleCarrito, ItemCarritoDetallado } from '../types/stock.types'
import Textarea from 'primevue/textarea'

const router = useRouter()
const authStore = useAuthStore()

const usuarioLogueado = computed(() => authStore.usuario)
const esOperario = computed(() => usuarioLogueado.value?.rol === 'operario')
const esVisualizador = computed(() => usuarioLogueado.value?.rol === 'visualizador')

/* ── Estados de Datos ── */
const carritos = ref<DetalleCarrito[]>([])
const carritoActivo = ref<DetalleCarrito | null>(null)
const herramientas = ref<any[]>([])
const isLoadingCarritos = ref(false)
const isLoadingItems = ref(false)
const errorCarga = ref('')

const dropdownAbierto = ref(false)
const procesandoID = ref<number | null>(null)
const menuNovedadAbierto = ref<number | null>(null)
const showModalResumen = ref(false)
const isSubmittingFirm = ref(false)
const submitSuccess = ref(false)

/* ── Modal Marca ── */
const modalMarcaAbierto = ref(false)
const herramientaParaMarca = ref<any>(null)
const busquedaMarcaModal = ref('')
const listadoMarcasAdmon = ref<string[]>([])
const isLoadingMarcas = ref(false)

const canvasRef = ref<HTMLCanvasElement | null>(null)
const context = ref<CanvasRenderingContext2D | null>(null)
const isDrawing = ref(false)

/* ── Filtros y búsqueda (deben declararse ANTES de herramientasFiltradas) ── */
const busqueda = ref('')
const filtroMarca = ref('Todas')
const filtroNovedad = ref('Todas')
const ordenCantidad = ref('default')

// Computed de herramientas filtradas - Usa las refs de arriba
const herramientasFiltradas = computed(() => {
  let filtradas = herramientas.value.filter((h: any) => !h.completado)
  if (busqueda.value) {
    const query = normalizar(busqueda.value)
    filtradas = filtradas.filter(h =>
      normalizar(h.nombre).includes(query) ||
      normalizar(h.nombreOriginal).includes(query) ||
      normalizar(h.codigo).includes(query) ||
      normalizar(h.marca).includes(query)
    )
  }
  if (filtroMarca.value !== 'Todas') filtradas = filtradas.filter(h => h.marca === filtroMarca.value)
  if (filtroNovedad.value !== 'Todas') filtradas = filtradas.filter(h => h.novedad === filtroNovedad.value)
  if (ordenCantidad.value === 'mayor-menor') filtradas.sort((a, b) => b.cantidad - a.cantidad)
  else if (ordenCantidad.value === 'menor-mayor') filtradas.sort((a, b) => a.cantidad - b.cantidad)
  return filtradas
})

/* ── Lazy Loading por Secciones ── */
const itemsAMostrar = ref(21) // Bloque inicial (7 filas de 3)
const pasoCarga = 12 // Cuántos sumamos al llegar al final
const centinela = ref<HTMLElement | null>(null)
let observador: IntersectionObserver | null = null

const herramientasVisibles = computed(() => {
  return herramientasFiltradas.value.slice(0, itemsAMostrar.value)
})

// Computed para saber si hay más items por cargar
const hayMasItems = computed(() => {
  return herramientasVisibles.value.length < herramientasFiltradas.value.length
})

function cargarMasSeccion(entries: IntersectionObserverEntry[]) {
  if (entries[0].isIntersecting && hayMasItems.value) {
    itemsAMostrar.value += pasoCarga
  }
}

// Configurar/Reconfigurar observador cuando el centinela cambia
function configurarObservador() {
  // Desconectar observador anterior si existe
  if (observador) {
    observador.disconnect()
    observador = null
  }

  // Si hay más items y el centinela existe, crear nuevo observador
  if (hayMasItems.value && centinela.value) {
    observador = new IntersectionObserver(cargarMasSeccion, {
      rootMargin: '200px', // Cargar antes de llegar al final
      threshold: 0.1
    })
    observador.observe(centinela.value)
  }
}

// Resetear scroll al buscar o filtrar
function resetearPaginacion() {
  itemsAMostrar.value = 21
}

const pendientes = computed(() => {
  if (!carritoActivo.value) return 0
  return (carritoActivo.value.registros || 0) - (carritoActivo.value.completados || 0)
})

const todoValidado = computed(() => {
  if (herramientas.value.length === 0) return false
  return herramientas.value.every(h => h.completado)
})

const itemsConDiscrepancia = computed(() => {
  return herramientas.value.filter(h => 
    h.novedad !== 'Sin novedad' || 
    h.cantidad !== h.cantidadOriginal ||
    h.marca !== h.marcaOriginal ||
    (h.observacion && h.observacion.trim().length > 0)
  )
})

const carritoYaFirmado = computed(() => {
  if (herramientas.value.length === 0) return false
  return herramientas.value.every(h => h.completadoBD)
})

async function cargarCarritos() {
  if (!authStore.usuario || !authStore.token) return
  
  isLoadingCarritos.value = true
  errorCarga.value = ''
  try {
    // Para visualizadores se pasa la cédula (empleado), para otros el id_usuario
    const cedula = esVisualizador.value ? authStore.usuario.empleado : undefined
    const res = await stockService.obtenerCarritosAsignados(authStore.usuario.id_usuario, authStore.token, cedula)
    carritos.value = res.carritos_asignados
    if (carritos.value.length > 0 && !carritoActivo.value) {
      await seleccionarCarrito(carritos.value[0])
    }
  } catch (err: any) {
    errorCarga.value = err.message
  } finally {
    isLoadingCarritos.value = false
  }
}

async function refrescarCarritos() {
  if (!authStore.usuario || !authStore.token) return
  isLoadingCarritos.value = true
  try {
    // Para visualizadores se pasa la cédula (empleado), para otros el id_usuario
    const cedula = esVisualizador.value ? authStore.usuario.empleado : undefined
    const res = await stockService.obtenerCarritosAsignados(authStore.usuario.id_usuario, authStore.token, cedula)
    carritos.value = res.carritos_asignados
  } catch (err: any) {
    console.error("Error validando novedades asíncronas:", err)
  } finally {
    isLoadingCarritos.value = false
  }
}

async function seleccionarCarrito(c: DetalleCarrito) {
  carritoActivo.value = c
  dropdownAbierto.value = false
  
  if (!authStore.usuario || !authStore.token) return
  
  isLoadingItems.value = true
  try {
    // Para visualizadores se pasa la cédula (empleado), para otros el id_usuario
    const cedula = esVisualizador.value ? authStore.usuario.empleado : undefined
    const res = await stockService.obtenerDetalladoCarrito(authStore.usuario.id_usuario, c.numero_carrito, authStore.token, cedula)
    // Mapeamos los ítems de SQL Server al formato esperado por el componente
    herramientas.value = res.items.map((item, idx) => ({
      id: idx + 1,
      codigo: item.referencia.trim(),
      nombre: item.nombre_inteligente || item.descripcion.trim(),
      nombreOriginal: item.descripcion.trim(),
      marca: item.ext1.trim() || 'SIN MARCA',
      marcaOriginal: item.ext1.trim() || 'SIN MARCA',
      cantidad: item.existencia,
      cantidadOriginal: item.existencia, // Para detectar discrepancias
      unidad: item.um.trim(),
      completado: item.completado === 1,
      completadoBD: item.completado === 1,
      // Campos de UI
      novedad: 'Sin novedad',
      resolucion: '',
      observacion: ''
    }))
  } catch (err: any) {
    console.error("Error al cargar detalle:", err)
  } finally {
    isLoadingItems.value = false
  }
}

function cerrarAlClickAfuera(e: MouseEvent) {
  const el = document.getElementById('cart-dropdown-root')
  if (el && !el.contains(e.target as Node)) dropdownAbierto.value = false
  if (menuAbierto.value) menuAbierto.value = null
  if (menuNovedadAbierto.value !== null) menuNovedadAbierto.value = null
}

onMounted(async () => {
  document.addEventListener('click', cerrarAlClickAfuera)
  await cargarCarritos()
})

// Watcher para configurar el observador cuando cambia el centinela o la disponibilidad de items
watch([centinela, hayMasItems], () => {
  configurarObservador()
}, { immediate: true, flush: 'post' })

onUnmounted(() => {
  document.removeEventListener('click', cerrarAlClickAfuera)
  if (observador) observador.disconnect()
})

/* herramientas ref ya se definió arriba */

/* ── Lógica Modal Marca ── */
async function abrirModalMarca(h: any) {
  herramientaParaMarca.value = h
  busquedaMarcaModal.value = ''
  
  // Abrir instantaneamente para no dar la impresion de espera
  modalMarcaAbierto.value = true
  
  if (listadoMarcasAdmon.value.length === 0) {
    isLoadingMarcas.value = true
    try {
      if (authStore.token) {
         const res = await stockService.obtenerListadoPartes(authStore.token)
         listadoMarcasAdmon.value = res.marcas || []
      }
    } catch(err: any) {
      console.error("Error cargando marcas: ", err.message)
    } finally {
      isLoadingMarcas.value = false
    }
  }
}

function seleccionarNuevaMarca(marca: string) {
  if (herramientaParaMarca.value) {
    herramientaParaMarca.value.marca = marca
  }
  modalMarcaAbierto.value = false
}

const marcasModalFiltradas = computed(() => {
  if (!busquedaMarcaModal.value) return listadoMarcasAdmon.value
  const q = normalizar(busquedaMarcaModal.value)
  return listadoMarcasAdmon.value.filter((m: string) => normalizar(m).includes(q))
})

const opcionesNovedad = ['Sin novedad', 'Dañada', 'Faltante']

// Escuchadores para resetear paginación
watch([busqueda, filtroMarca, filtroNovedad, ordenCantidad], () => {
  resetearPaginacion()
})

/* ── Lógica de Firma ── */
function initCanvas() {
  if (!canvasRef.value) return
  context.value = canvasRef.value.getContext('2d')
  if (context.value) {
    context.value.strokeStyle = '#0f172a'
    context.value.lineWidth = 3
    context.value.lineCap = 'round'
    context.value.lineJoin = 'round'
  }
}

function startDrawing(e: MouseEvent | TouchEvent) {
  isDrawing.value = true
  const pos = getPos(e)
  context.value?.beginPath()
  context.value?.moveTo(pos.x, pos.y)
}

function draw(e: MouseEvent | TouchEvent) {
  if (!isDrawing.value) return
  const pos = getPos(e)
  context.value?.lineTo(pos.x, pos.y)
  context.value?.stroke()
}

function stopDrawing() {
  isDrawing.value = false
}

function clearSignature() {
  if (!canvasRef.value || !context.value) return
  context.value.clearRect(0, 0, canvasRef.value.width, canvasRef.value.height)
}

function getPos(e: MouseEvent | TouchEvent) {
  const canvas = canvasRef.value
  if (!canvas) return { x: 0, y: 0 }
  const rect = canvas.getBoundingClientRect()
  const clientX = 'touches' in e ? e.touches[0].clientX : e.clientX
  const clientY = 'touches' in e ? e.touches[0].clientY : e.clientY
  return {
    x: clientX - rect.left,
    y: clientY - rect.top
  }
}

async function confirmarYEnviar() {
  if (!authStore.usuario?.id_usuario || !authStore.token || !carritoActivo.value || !canvasRef.value) return
  
  const blank = document.createElement('canvas')
  blank.width = canvasRef.value.width
  blank.height = canvasRef.value.height
  if (canvasRef.value.toDataURL() === blank.toDataURL()) {
    alert("Es obligatorio dibujar una firma para continuar.")
    return
  }
  
  const firmaBase64 = canvasRef.value.toDataURL('image/png')
  
  const registros = herramientas.value.map((h: any) => ({
    id_usuario: authStore.usuario!.id_usuario,
    numero_carrito: carritoActivo.value!.numero_carrito,
    nombre_carrito: `Carrito ${carritoActivo.value!.numero_carrito}`,
    id_producto: h.codigo,
    referencia_producto: h.codigo,
    descripcion_producto: h.nombre,
    marca_adicional: h.marcaOriginal,
    marca: h.marca,
    cantidad_sistema: h.cantidadOriginal,
    cantidad_fisica: h.cantidad,
    unidad_medida: h.unidad,
    novedad: h.novedad === 'Sin novedad' ? 'ninguna' : (h.novedad === 'Dañada' ? 'desgaste' : 'faltante'),
    accion_faltante: h.resolucion ? h.resolucion.toLowerCase() : 'ninguna',
    observacion: h.observacion,
    firma_digital: firmaBase64.split(',')[1] // Solo el base64 sin el prefijo data:image/png;base64,
  }))

  isSubmittingFirm.value = true
  submitSuccess.value = false

  try {
    await stockService.guardarInventario(registros, authStore.token)
    isSubmittingFirm.value = false
    submitSuccess.value = true
    
    setTimeout(async () => {
      showModalResumen.value = false
      submitSuccess.value = false
      clearSignature()
      await cargarCarritos()
      // Recargar el detalle para que carritoYaFirmado se actualice desde el backend
      if (carritoActivo.value) {
        await seleccionarCarrito(carritoActivo.value)
      }
    }, 2500)
    
  } catch (err: any) {
    alert("Error al guardar final: " + err.message)
    isSubmittingFirm.value = false
  }
}

function toggleMenu(nombre: string) {
  if (menuAbierto.value === nombre) menuAbierto.value = null
  else menuAbierto.value = nombre
}

function toggleNovedadMenu(id: number) {
  if (menuNovedadAbierto.value === id) menuNovedadAbierto.value = null
  else menuNovedadAbierto.value = id
}

const marcasUnicas = computed(() => {
  return ['Todas', ...new Set(herramientas.value.map(h => h.marca))].sort()
})

const noveladesUnicas = computed(() => ['Todas', ...opcionesNovedad])

function normalizar(texto: string) {
  return texto?.normalize("NFD").replace(/[\u0300-\u036f]/g, "").toLowerCase() || ""
}

function resaltarTexto(texto: string, consulta: string) {
  if (!consulta) return texto
  const textoNorm = normalizar(texto)
  const consultaNorm = normalizar(consulta)
  let resultado = ""
  let posActual = 0
  let indice = textoNorm.indexOf(consultaNorm)
  if (indice === -1) return texto
  while (indice !== -1) {
    resultado += texto.substring(posActual, indice)
    resultado += `<mark class="bg-green-100 text-green-900 px-0.5 rounded-sm border-b-2 border-green-300 font-bold">${texto.substring(indice, indice + consulta.length)}</mark>`
    posActual = indice + consulta.length
    indice = textoNorm.indexOf(consultaNorm, posActual)
  }
  resultado += texto.substring(posActual)
  return resultado
}

function decrementar(h: { cantidad: number }) { if (h.cantidad > 0) h.cantidad-- }
function incrementar(h: { cantidad: number }) { h.cantidad++ }

async function guardarCambios(h: any) {
  if (!carritoActivo.value) return
  
  procesandoID.value = h.id
  
  // Retraso muy leve para dar retroalimentación visual de procesamiento fluido
  await new Promise(resolve => setTimeout(resolve, 350))
  
  // Marcamos la tarjeta como completada a nivel lógico
  h.completado = true
  
  // Actualizamos el sumatorio de partes completadas directamente en la memoria
  // para que la interfaz siga funcionando sin golpear la base de datos
  const carritoLocal = carritos.value.find(c => c.numero_carrito === carritoActivo.value?.numero_carrito)
  if (carritoLocal && carritoLocal.partes_completadas < carritoLocal.total_partes) {
    carritoLocal.partes_completadas++
  }
  
  procesandoID.value = null
}

function logout() {
  authStore.cerrarSesion()
  router.push('/login')
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
          <span class="text-[0.6rem] font-semibold text-slate-500 tracking-widest uppercase">CARRITO DE MANTENIMIENTO</span>
        </div>
      </div>

      <div class="flex items-center gap-6">
        <nav v-if="!esOperario && !esVisualizador" class="hidden md:flex items-center gap-1 bg-slate-100/80 p-1 rounded-full border border-slate-200">
          <router-link to="/dashboard/stock" class="px-4 py-1.5 text-xs font-bold bg-white text-green-700 rounded-full shadow-sm border border-slate-200 pointer-events-none">Stock</router-link>
          <router-link to="/dashboard/prestamo" class="px-4 py-1.5 text-xs font-bold text-slate-500 rounded-full transition-all hover:text-slate-800">Préstamo</router-link>
          <router-link to="/dashboard/admin" class="px-4 py-1.5 text-xs font-bold text-slate-500 rounded-full transition-all hover:text-slate-800">Admin</router-link>
        </nav>
        <div v-if="esVisualizador" class="hidden sm:flex items-center gap-2 px-3 py-1.5 bg-blue-50 border border-blue-200 rounded-full">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="#3b82f6" stroke-width="2">
            <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"></path>
            <circle cx="12" cy="12" r="3"></circle>
          </svg>
          <span class="text-xs font-bold text-blue-600">Modo Visualización</span>
        </div>
        <div class="hidden sm:flex flex-col items-end leading-none gap-0.5 pl-4 border-l border-slate-200">
          <span class="text-[0.82rem] font-extrabold text-slate-800 tracking-wide uppercase">{{ usuarioLogueado?.nombre_completo || 'USUARIO' }}</span>
          <span class="text-[0.6rem] text-slate-500 tracking-widest uppercase">{{ usuarioLogueado?.rol || 'SIN ROL' }}</span>
        </div>
        <button @click="logout" class="flex items-center justify-center w-10 h-10 rounded-full bg-white border border-slate-200 text-slate-500 shadow-sm transition-all duration-200 hover:bg-red-50 hover:text-red-500 hover:border-red-200 cursor-pointer">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"></path>
            <polyline points="16 17 21 12 16 7"></polyline>
            <line x1="21" y1="12" x2="9" y2="12"></line>
          </svg>
        </button>
      </div>
    </header>

    <!-- ── Contenido principal ── -->
    <main class="max-w-6xl mx-auto px-4 py-10 flex flex-col items-center">

      <!-- Selector de carrito -->
      <div id="cart-dropdown-root" class="relative w-full max-w-xs mb-8">
        <button
          @click.stop="dropdownAbierto = !dropdownAbierto; if(dropdownAbierto) refrescarCarritos()"
          :class="['w-full flex items-center justify-between px-5 py-3.5 text-sm font-bold rounded-2xl shadow-sm transition-all duration-200 cursor-pointer', dropdownAbierto ? 'bg-green-50 border-2 border-green-400 text-green-700' : 'bg-white border border-slate-200 hover:shadow-md text-slate-700']"
        >
          <span class="flex items-center gap-2.5">
            <svg width="17" height="17" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="2" y="7" width="20" height="14" rx="2" ry="2"></rect><path d="M16 21V5a2 2 0 0 0-2-2h-4a2 2 0 0 0-2 2v16"></path></svg>
            {{ carritoActivo ? `Carrito ${carritoActivo?.numero_carrito}` : 'Seleccionar Carrito' }}
          </span>
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" class="transition-transform duration-300" :class="dropdownAbierto ? 'rotate-180 text-green-500' : ''">
            <polyline points="6 9 12 15 18 9"></polyline>
          </svg>
        </button>
      </div>

      <!-- Modal de selección de carrito -->
      <Teleport to="body">
        <Transition enter-active-class="transition-opacity duration-300 ease-out" enter-from-class="opacity-0" enter-to-class="opacity-100" leave-active-class="transition-opacity duration-200 ease-in" leave-from-class="opacity-100" leave-to-class="opacity-0">
          <div v-if="dropdownAbierto" class="fixed inset-0 z-[100] flex items-center justify-center px-4">
            <div class="absolute inset-0" style="background: rgba(15, 23, 42, 0.45); backdrop-filter: blur(2px);" @click="dropdownAbierto = false"></div>
            <Transition appear enter-active-class="transition-all duration-[450ms] ease-[cubic-bezier(0.34,1.56,0.64,1)]" enter-from-class="opacity-0 scale-75 translate-y-12" enter-to-class="opacity-100 scale-100 translate-y-0" leave-active-class="transition-all duration-200 ease-in" leave-from-class="opacity-100 scale-100 translate-y-0" leave-to-class="opacity-0 scale-95 translate-y-4">
              <div v-if="dropdownAbierto" class="relative bg-white rounded-3xl shadow-2xl w-full max-w-md overflow-hidden z-10" @click.stop>
                <div class="flex items-center gap-3.5 px-6 pt-6 pb-4 border-b border-slate-100/80">
                  <div class="w-11 h-11 rounded-2xl bg-green-50 border border-green-100 flex items-center justify-center shrink-0">
                    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="#22c55e" stroke-width="2"><rect x="2" y="7" width="20" height="14" rx="2" ry="2"></rect><path d="M16 21V5a2 2 0 0 0-2-2h-4a2 2 0 0 0-2 2v16"></path></svg>
                  </div>
                  <div class="flex-1 min-w-0">
                    <h2 class="text-base font-extrabold text-slate-800 leading-tight">Seleccionar Carrito</h2>
                    <p class="text-xs text-slate-400 font-medium mt-0.5">{{ carritos.length }} carritos disponibles</p>
                  </div>
                  <button @click="dropdownAbierto = false" class="shrink-0 flex items-center justify-center w-8 h-8 rounded-full text-slate-400 hover:bg-red-50 hover:text-red-400 transition-all duration-200 cursor-pointer border border-transparent hover:border-red-100">
                    <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><line x1="18" y1="6" x2="6" y2="18"></line><line x1="6" y1="6" x2="18" y2="18"></line></svg>
                  </button>
                </div>
                <ul class="divide-y divide-slate-50 max-h-72 overflow-y-auto">
                  <li v-for="c in carritos" :key="c.numero_carrito" @click="seleccionarCarrito(c)" :class="['flex items-center gap-4 px-6 py-4 cursor-pointer transition-all duration-150 group', carritoActivo?.numero_carrito === c.numero_carrito ? 'bg-green-50' : 'hover:bg-slate-50']">
                    <span :class="['w-3 h-3 rounded-full shrink-0 transition-all duration-200', carritoActivo?.numero_carrito === c.numero_carrito ? 'bg-green-500 shadow-[0_0_0_3px_rgba(34,197,94,0.22)]' : 'bg-slate-200 group-hover:bg-slate-300']"></span>
                    <div class="flex-1 min-w-0">
                      <p :class="['text-sm font-bold leading-tight', carritoActivo?.numero_carrito === c?.numero_carrito ? 'text-green-600' : 'text-slate-800']">Carrito {{ c?.numero_carrito }}</p>
                      <p v-if="c?.nombre" class="text-xs text-slate-400 font-medium mt-0.5 truncate">{{ c?.nombre }}</p>
                    </div>
                    <div class="flex flex-col items-end gap-1">
                      <span :class="['text-[0.65rem] font-extrabold rounded-full px-3 py-1 shrink-0 whitespace-nowrap tracking-wide transition-colors', carritoActivo?.numero_carrito === c?.numero_carrito ? 'bg-green-100 text-green-600' : 'bg-slate-100 text-slate-400']">{{ c?.registros }} ITEMS</span>
                      <span v-if="c?.completados > 0" class="text-[0.55rem] font-bold text-green-500 tracking-tighter">{{ c?.completados }} OK</span>
                    </div>
                    <svg v-if="carritoActivo?.numero_carrito === c.numero_carrito" width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="#22c55e" stroke-width="2.5" class="shrink-0"><polyline points="20 6 9 17 4 12"></polyline></svg>
                  </li>
                </ul>
                <div class="px-6 py-3.5 border-t border-slate-100 flex items-center justify-between bg-slate-50/60">
                  <p class="text-xs text-slate-400 font-medium">Activo: <span class="font-extrabold text-green-600">{{ carritoActivo ? `Carrito ${carritoActivo?.numero_carrito}` : 'Ninguno' }}</span></p>
                  <button @click="dropdownAbierto = false" class="text-xs font-bold text-slate-500 hover:text-slate-800 transition-colors cursor-pointer">Cerrar</button>
                </div>
              </div>
            </Transition>
          </div>
        </Transition>
      </Teleport>

      <div v-if="carritoActivo" class="text-center flex flex-col items-center gap-3 mb-10">
        <h1 class="text-5xl font-black tracking-tight leading-tight" style="background: linear-gradient(135deg, #0f172a 0%, #334155 100%); -webkit-background-clip: text; -webkit-text-fill-color: transparent; background-clip: text;">
          Carrito {{ carritoActivo?.numero_carrito }}
        </h1>
        <div v-if="carritoActivo?.nombre" class="inline-flex items-center gap-2 text-sm font-semibold text-slate-500 bg-white border border-slate-200 rounded-full px-4 py-2 shadow-sm">
          <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"></path><circle cx="12" cy="7" r="4"></circle></svg>
          Responsable: {{ carritoActivo?.nombre }}
        </div>
        <div class="flex items-center gap-2 text-sm">
          <span class="font-medium text-slate-500">{{ esVisualizador ? 'Visualiza tus herramientas.' : 'Gestiona tus herramientas.' }}</span>
          <span v-if="!esVisualizador" class="bg-green-100 text-green-700 text-xs font-extrabold border border-green-200 rounded-full px-3 py-1 badge-pulse">{{ pendientes }} Pendientes</span>
        </div>
      </div>

      <!-- Barra de búsqueda -->
      <div class="w-full max-w-2xl mb-8 relative group">
        <div class="absolute inset-y-0 left-5 flex items-center pointer-events-none">
          <svg class="text-slate-400 group-focus-within:text-green-500 transition-colors" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
            <circle cx="11" cy="11" r="8"></circle><line x1="21" y1="21" x2="16.65" y2="16.65"></line>
          </svg>
        </div>
        <input
          v-model="busqueda"
          type="text"
          placeholder="Buscar herramienta por nombre, código o marca..."
          class="w-full py-4 pl-14 pr-32 bg-white border-2 border-slate-200 rounded-full text-slate-700 font-semibold shadow-sm focus:outline-none focus:border-green-400 focus:ring-4 focus:ring-green-100 transition-all placeholder:text-slate-300 placeholder:font-medium"
        />
        <div class="absolute inset-y-0 right-3 flex items-center">
          <span class="bg-slate-100 text-slate-400 text-[0.65rem] font-bold px-3 py-1 rounded-full uppercase tracking-wider">{{ herramientasFiltradas.length }} items</span>
        </div>
      </div>

      <!-- Barra de Filtros Avanzados -->
      <div class="w-full max-w-5xl mb-10 flex flex-wrap items-center justify-center gap-4 bg-white p-5 rounded-[2.5rem] border border-slate-100 shadow-xl shadow-slate-200/40 relative z-40">

        <!-- Filtro Marca -->
        <div class="flex flex-col gap-1.5 px-4 border-r border-slate-100 last:border-none relative">
          <label class="text-[0.62rem] font-black text-slate-400 uppercase tracking-widest pl-1">Fabricante</label>
          <div @click.stop="toggleMenu('marca')" class="bg-slate-50 text-[0.75rem] font-extrabold text-slate-700 px-4 py-2.5 rounded-xl flex items-center justify-between gap-3 cursor-pointer min-w-[140px] hover:bg-slate-100 transition-colors border border-slate-200/50">
            <span class="truncate">{{ filtroMarca }}</span>
            <svg class="transition-transform duration-300" :class="{'rotate-180': menuAbierto === 'marca'}" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3"><polyline points="6 9 12 15 18 9"></polyline></svg>
          </div>
          <Transition name="fade-slide">
            <div v-if="menuAbierto === 'marca'" class="absolute top-[110%] left-0 w-full min-w-[180px] bg-white border border-slate-100 shadow-2xl rounded-2xl p-2 z-[60]">
              <div v-for="m in marcasUnicas" :key="m" @click="filtroMarca = m; menuAbierto = null" :class="['px-4 py-2.5 text-xs font-bold rounded-xl cursor-pointer transition-all flex items-center justify-between', filtroMarca === m ? 'bg-green-50 text-green-700' : 'text-slate-600 hover:bg-slate-50']">
                {{ m }}
                <svg v-if="filtroMarca === m" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="4"><polyline points="20 6 9 17 4 12"></polyline></svg>
              </div>
            </div>
          </Transition>
        </div>

        <!-- Filtro Novedad -->
        <div class="flex flex-col gap-1.5 px-4 border-r border-slate-100 last:border-none relative">
          <label class="text-[0.62rem] font-black text-slate-400 uppercase tracking-widest pl-1">Novedad</label>
          <div @click.stop="toggleMenu('novedad')" class="bg-slate-50 text-[0.75rem] font-extrabold text-slate-700 px-4 py-2.5 rounded-xl flex items-center justify-between gap-3 cursor-pointer min-w-[140px] hover:bg-slate-100 transition-colors border border-slate-200/50">
            <span class="truncate">{{ filtroNovedad }}</span>
            <svg class="transition-transform duration-300" :class="{'rotate-180': menuAbierto === 'novedad'}" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3"><polyline points="6 9 12 15 18 9"></polyline></svg>
          </div>
          <Transition name="fade-slide">
            <div v-if="menuAbierto === 'novedad'" class="absolute top-[110%] left-0 w-full min-w-[180px] bg-white border border-slate-100 shadow-2xl rounded-2xl p-2 z-[60]">
              <div v-for="n in noveladesUnicas" :key="n" @click="filtroNovedad = n; menuAbierto = null" :class="['px-4 py-2.5 text-xs font-bold rounded-xl cursor-pointer transition-all flex items-center justify-between', filtroNovedad === n ? 'bg-green-50 text-green-700' : 'text-slate-600 hover:bg-slate-50']">
                {{ n }}
                <svg v-if="filtroNovedad === n" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="4"><polyline points="20 6 9 17 4 12"></polyline></svg>
              </div>
            </div>
          </Transition>
        </div>

        <!-- Ordenamiento -->
        <div class="flex flex-col gap-1.5 px-4 border-r border-slate-100 last:border-none relative">
          <label class="text-[0.62rem] font-black text-slate-400 uppercase tracking-widest pl-1">Ordenar por</label>
          <div @click.stop="toggleMenu('orden')" class="bg-slate-50 text-[0.75rem] font-extrabold text-slate-700 px-4 py-2.5 rounded-xl flex items-center justify-between gap-3 cursor-pointer min-w-[160px] hover:bg-slate-100 transition-colors border border-slate-200/50">
            <span class="truncate">{{ ordenCantidad === 'default' ? 'Nombre (A-Z)' : ordenCantidad === 'mayor-menor' ? 'Cantidad: Mayor' : 'Cantidad: Menor' }}</span>
            <svg class="transition-transform duration-300" :class="{'rotate-180': menuAbierto === 'orden'}" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3"><polyline points="6 9 12 15 18 9"></polyline></svg>
          </div>
          <Transition name="fade-slide">
            <div v-if="menuAbierto === 'orden'" class="absolute top-[110%] left-0 w-full min-w-[200px] bg-white border border-slate-100 shadow-2xl rounded-2xl p-2 z-[60]">
              <div @click="ordenCantidad = 'default'; menuAbierto = null" :class="['px-4 py-2.5 text-xs font-bold rounded-xl cursor-pointer transition-all flex items-center justify-between', ordenCantidad === 'default' ? 'bg-green-50 text-green-700' : 'text-slate-600 hover:bg-slate-50']">
                Nombre (A-Z) <svg v-if="ordenCantidad === 'default'" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="4"><polyline points="20 6 9 17 4 12"></polyline></svg>
              </div>
              <div @click="ordenCantidad = 'mayor-menor'; menuAbierto = null" :class="['px-4 py-2.5 text-xs font-bold rounded-xl cursor-pointer transition-all flex items-center justify-between', ordenCantidad === 'mayor-menor' ? 'bg-green-50 text-green-700' : 'text-slate-600 hover:bg-slate-50']">
                Cantidad: Mayor a Menor <svg v-if="ordenCantidad === 'mayor-menor'" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="4"><polyline points="20 6 9 17 4 12"></polyline></svg>
              </div>
              <div @click="ordenCantidad = 'menor-mayor'; menuAbierto = null" :class="['px-4 py-2.5 text-xs font-bold rounded-xl cursor-pointer transition-all flex items-center justify-between', ordenCantidad === 'menor-mayor' ? 'bg-green-50 text-green-700' : 'text-slate-600 hover:bg-slate-50']">
                Cantidad: Menor a Mayor <svg v-if="ordenCantidad === 'menor-mayor'" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="4"><polyline points="20 6 9 17 4 12"></polyline></svg>
              </div>
            </div>
          </Transition>
        </div>

        <!-- Botón Reset -->
        <button @click="filtroMarca = 'Todas'; filtroNovedad = 'Todas'; ordenCantidad = 'default'; busqueda = ''; menuAbierto = null" class="ml-2 w-11 h-11 rounded-2xl bg-white border border-slate-100 flex items-center justify-center text-slate-400 hover:bg-red-50 hover:text-red-500 hover:border-red-100 transition-all shadow-sm group" title="Limpiar Filtros">
          <svg class="group-hover:rotate-90 transition-transform duration-300" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round"><path d="M18 6L6 18M6 6l12 12"></path></svg>
        </button>
      </div>



      <!-- Skeleton Loader (Estado de Carga) -->
      <div v-if="isLoadingItems" class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-5 w-full">
        <div v-for="i in 6" :key="i" class="bg-white border border-slate-100 rounded-3xl p-6 flex flex-col gap-5 shadow-sm animate-pulse">
          <div class="flex items-start gap-4">
            <div class="w-14 h-14 bg-slate-100 rounded-2xl shrink-0"></div>
            <div class="flex-1 flex flex-col gap-2 pt-1">
              <div class="h-4 bg-slate-100 rounded-md w-3/4"></div>
              <div class="h-3 bg-slate-100 rounded-md w-1/2"></div>
            </div>
          </div>
          <div class="flex flex-col gap-4">
            <div class="h-10 bg-slate-50 rounded-xl w-full"></div>
            <div class="flex items-center justify-between px-4">
              <div class="w-10 h-10 bg-slate-100 rounded-xl"></div>
              <div class="h-8 bg-slate-100 rounded-md w-12"></div>
              <div class="w-10 h-10 bg-slate-100 rounded-xl"></div>
            </div>
            <div class="h-20 bg-slate-50 rounded-2xl w-full"></div>
            <div class="h-12 bg-slate-200 rounded-2xl w-full mt-2"></div>
          </div>
        </div>
      </div>

      <!-- Grid de tarjetas real -->
      <TransitionGroup 
        v-else
        name="staggered-fade" 
        tag="section" 
        class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-5 w-full items-start"
      >
        <div
          v-for="(h, index) in herramientasVisibles"
          :key="h.id"
          class="card-animated relative bg-white border border-slate-200/80 rounded-3xl p-6 flex flex-col gap-5 overflow-visible shadow-sm hover:shadow-xl hover:-translate-y-1 transition-all duration-500 group"
          :style="`animation-delay: ${Math.min(index * 0.08, 1.2)}s`"
        >
          <!-- Contenedor de recortes para decoraciones -->
          <div class="absolute inset-0 rounded-3xl overflow-hidden pointer-events-none">
            <div class="absolute -top-9 -right-9 w-28 h-28 bg-green-50 rounded-full opacity-60 group-hover:opacity-100 transition-opacity duration-300"></div>
          </div>

          <!-- Loader Overlay (Individual por Card) -->
          <Transition
            enter-active-class="transition duration-300 ease-out"
            enter-from-class="opacity-0"
            enter-to-class="opacity-100"
            leave-active-class="transition duration-200 ease-in"
            leave-from-class="opacity-100"
            leave-to-class="opacity-0"
          >
            <div 
              v-if="procesandoID === h.id" 
              class="absolute inset-0 z-[60] bg-white/95 rounded-3xl flex flex-col items-center justify-center p-6 text-center"
            >
              <div class="relative w-12 h-12 mb-3">
                <div class="absolute inset-0 border-4 border-green-100 rounded-full"></div>
                <div class="absolute inset-0 border-4 border-green-500 rounded-full border-t-transparent animate-spin"></div>
              </div>
              <p class="text-[0.65rem] font-black text-green-700 uppercase tracking-[0.2em] animate-pulse">Guardando cambios...</p>
            </div>
          </Transition>

          <!-- Cabecera tarjeta -->
          <div class="flex items-start gap-4 relative z-10">
            <div class="w-14 h-14 shrink-0 flex items-center justify-center bg-white border border-slate-100 rounded-2xl shadow-sm">
              <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="#22c55e" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                <path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"></path>
                <polyline points="3.27 6.96 12 12.01 20.73 6.96"></polyline>
                <line x1="12" y1="22.08" x2="12" y2="12"></line>
              </svg>
            </div>
            <div class="flex flex-col gap-1 pt-0.5 relative">
              <h3 class="text-base font-extrabold text-slate-800 leading-snug group-hover:text-green-600 transition-colors" v-html="resaltarTexto(h.nombre, busqueda)"></h3>
              <div class="flex flex-col gap-0.5">
                <p class="text-xs font-bold text-slate-400 tracking-wide" v-html="resaltarTexto(h.codigo, busqueda)"></p>
                <p v-if="h.nombreOriginal && h.nombre !== h.nombreOriginal" 
                   class="text-[0.6rem] font-bold text-slate-300 uppercase truncate"
                   v-html="resaltarTexto(h.nombreOriginal, busqueda)"></p>
              </div>
              <!-- Indicador de Completado -->
              <div v-if="h.completado" class="absolute -top-1 -right-2 bg-green-500 text-white rounded-full p-1 shadow-lg ring-2 ring-white">
                <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="4"><polyline points="20 6 9 17 4 12"></polyline></svg>
              </div>
            </div>
          </div>

          <!-- Cuerpo -->
          <div class="flex flex-col gap-4 flex-1 relative z-10">
            <!-- Marca -->
            <div class="flex flex-col gap-1">
              <div class="flex justify-between items-center">
                <label class="text-[0.62rem] font-extrabold text-slate-400 tracking-widest uppercase">MARCA / FABRICANTE</label>
                <Transition enter-active-class="transition-all duration-300 ease-out" enter-from-class="opacity-0 scale-90 translate-x-2" enter-to-class="opacity-100 scale-100 translate-x-0" leave-active-class="transition-all duration-200 ease-in" leave-from-class="opacity-100 scale-100 translate-x-0" leave-to-class="opacity-0 scale-90 translate-x-2">
                  <span v-if="h.marca !== h.marcaOriginal" class="bg-[#4cc253]/15 text-[#4cc253] text-[0.55rem] font-black px-2 py-0.5 rounded-md uppercase tracking-wider border border-[#4cc253]/30 shadow-sm">MODIFICADO</span>
                </Transition>
              </div>
              <div @click="!esVisualizador && abrirModalMarca(h)" :class="['rounded-xl px-4 py-2.5 text-sm font-extrabold transition-colors flex justify-between items-center border shadow-sm', h.marca !== h.marcaOriginal ? 'bg-[#4cc253]/5 border-[#4cc253]/40 text-slate-800' : 'bg-slate-50 border-slate-100 text-slate-700', esVisualizador ? '' : 'cursor-pointer hover:bg-slate-100 hover:border-slate-300 group/marca']">
                <span class="truncate" v-html="resaltarTexto(h.marca, busqueda)"></span>
                <div v-if="!esVisualizador" :class="['text-[0.6rem] font-bold transition-all flex items-center gap-1 shrink-0', h.marca !== h.marcaOriginal ? 'text-[#4cc253] group-hover/marca:scale-110' : 'text-slate-400 group-hover/marca:text-[#4cc253]']">
                  <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><path d="M12 20h9"></path><path d="M16.5 3.5a2.121 2.121 0 0 1 3 3L7 19l-4 1 1-4L16.5 3.5z"></path></svg>
                </div>
              </div>
            </div>

            <!-- Cantidad -->
            <div class="flex items-center justify-between">
              <button v-if="!esVisualizador" @click="decrementar(h)" class="w-11 h-11 flex items-center justify-center bg-white border-2 border-slate-200 rounded-2xl text-slate-500 shadow-sm transition-all duration-200 hover:border-green-400 hover:text-green-600 hover:bg-green-50 hover:scale-105 active:scale-95 cursor-pointer">
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><line x1="5" y1="12" x2="19" y2="12"></line></svg>
              </button>
              <div v-else class="w-11"></div>
              <div class="flex flex-col items-center">
                <span class="text-4xl font-black text-slate-900 leading-none tabular-nums">{{ h.cantidad }}</span>
                <span class="text-[0.58rem] font-black text-slate-400 tracking-widest mt-1 uppercase">UND</span>
              </div>
              <button v-if="!esVisualizador" @click="incrementar(h)" class="w-11 h-11 flex items-center justify-center bg-white border-2 border-slate-200 rounded-2xl text-slate-500 shadow-sm transition-all duration-200 hover:border-green-400 hover:text-green-600 hover:bg-green-50 hover:scale-105 active:scale-95 cursor-pointer">
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><line x1="12" y1="5" x2="12" y2="19"></line><line x1="5" y1="12" x2="19" y2="12"></line></svg>
              </button>
              <div v-else class="w-11"></div>
            </div>

            <hr class="border-slate-100" />

            <!-- Novedad (Solo lectura para visualizadores) -->
            <div v-if="!esVisualizador" class="flex flex-col gap-1 relative">
              <label class="text-[0.62rem] font-extrabold text-slate-400 tracking-widest uppercase">NOVEDAD</label>
              <div
                @click.stop="toggleNovedadMenu(h.id)"
                :class="['flex items-center justify-between px-4 py-2.5 rounded-xl text-sm font-bold cursor-pointer transition-all border',
                  h.novedad === 'Sin novedad' ? 'bg-slate-50 border-slate-200/60 text-slate-700 hover:bg-slate-100'
                  : h.novedad === 'Dañada' ? 'bg-orange-50 border-orange-200 text-orange-700'
                  : 'bg-red-50 border-red-200 text-red-700']"
              >
                <span>{{ h.novedad }}</span>
                <svg class="transition-transform duration-300" :class="{'rotate-180': menuNovedadAbierto === h.id}" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3"><polyline points="6 9 12 15 18 9"></polyline></svg>
              </div>
              <Transition name="fade-slide">
                <div v-if="menuNovedadAbierto === h.id" class="absolute top-[110%] left-0 right-0 bg-white border border-slate-100 shadow-2xl rounded-2xl p-2 z-[50]">
                  <div v-for="op in opcionesNovedad" :key="op" @click.stop="h.novedad = op; menuNovedadAbierto = null" :class="['px-4 py-2.5 text-xs font-bold rounded-xl cursor-pointer transition-all flex items-center justify-between', h.novedad === op ? 'bg-green-50 text-green-700' : 'text-slate-600 hover:bg-slate-50']">
                    {{ op }}
                    <svg v-if="h.novedad === op" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="4"><polyline points="20 6 9 17 4 12"></polyline></svg>
                  </div>
                </div>
              </Transition>
            </div>

            <!-- Resolución de Faltante (Solo si novedad es Faltante y no es visualizador) -->
            <Transition
              v-if="!esVisualizador"
              enter-active-class="transition-all duration-300 ease-out"
              enter-from-class="opacity-0 -translate-y-2 scale-95"
              enter-to-class="opacity-100 translate-y-0 scale-100"
              leave-active-class="transition-all duration-200 ease-in"
              leave-from-class="opacity-100 translate-y-0 scale-100"
              leave-to-class="opacity-0 -translate-y-2 scale-95"
            >
              <div v-if="h.novedad === 'Faltante'" class="flex flex-col gap-1.5 p-3.5 bg-red-50/50 border border-red-100 rounded-2xl">
                <label class="text-[0.6rem] font-black text-red-400 uppercase tracking-widest pl-1">Seleccionar Resolución</label>
                <div class="flex gap-2">
                  <button 
                    @click="h.resolucion = 'Descuento'"
                    :class="['flex-1 py-2 text-[0.65rem] font-black rounded-xl transition-all border', h.resolucion === 'Descuento' ? 'bg-red-500 text-white border-red-600 shadow-md translate-y-[-1px]' : 'bg-white text-red-500 border-red-100 hover:bg-red-50']"
                  >DESCUENTO</button>
                  <button 
                    @click="h.resolucion = 'Compra'"
                    :class="['flex-1 py-2 text-[0.65rem] font-black rounded-xl transition-all border', h.resolucion === 'Compra' ? 'bg-red-500 text-white border-red-600 shadow-md translate-y-[-1px]' : 'bg-white text-red-500 border-red-100 hover:bg-red-50']"
                  >COMPRA</button>
                </div>
              </div>
            </Transition>

            <!-- Observación (Solo para no visualizadores) -->
            <div v-if="!esVisualizador" class="flex flex-col gap-1 pv-field">
              <label class="text-[0.62rem] font-extrabold text-slate-400 tracking-widest uppercase">OBSERVACIÓN</label>
              <Textarea v-model="h.observacion" rows="2" placeholder="Justifique el cambio..." autoResize fluid />
            </div>
          </div>

          <!-- Botón guardar (Solo para no visualizadores) -->
          <button 
            v-if="!esVisualizador"
            @click="guardarCambios(h)"
            class="w-full flex items-center justify-center gap-2 py-3.5 rounded-2xl text-white font-black text-[0.65rem] uppercase tracking-wider cursor-pointer transition-colors bg-[#4cc253] hover:bg-[#43ad49] border border-transparent hover:border-[#38913d]"
          >
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
              <path d="M19 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11l5 5v11a2 2 0 0 1-2 2z"></path>
              <polyline points="17 21 17 13 7 13 7 21"></polyline>
              <polyline points="7 3 7 8 15 8"></polyline>
            </svg>
            Guardar Cambios
          </button>
        </div>
      </TransitionGroup>

      <!-- Centinela para lazy loading -->
      <div 
        v-if="herramientasVisibles.length < herramientasFiltradas.length"
        ref="centinela" 
        class="w-full h-20 flex items-center justify-center"
      >
        <div class="animate-pulse flex items-center gap-2 text-slate-400">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="12" cy="12" r="10" stroke-dasharray="60" stroke-dashoffset="20" class="animate-spin" style="animation-duration: 1s;"></circle>
          </svg>
          <span class="text-xs font-semibold">Cargando más...</span>
        </div>
      </div>

      <!-- Botón Flotante de Finalización (Solo si 100% completado, no firmado en BD, y no es visualizador) -->
      <Transition
        enter-active-class="transition-all duration-300 ease-out"
        enter-from-class="opacity-0 translate-y-10"
        enter-to-class="opacity-100 translate-y-0"
        leave-active-class="transition-all duration-200 ease-in"
        leave-from-class="opacity-100 translate-y-0"
        leave-to-class="opacity-0 translate-y-10"
      >
        <button 
          v-if="!esVisualizador && todoValidado && !carritoYaFirmado"
          @click="showModalResumen = true; nextTick(() => initCanvas())"
          class="fixed bottom-10 right-10 z-[80] flex items-center gap-2.5 px-7 py-4 bg-[#4cc253] hover:bg-[#43ad49] text-white rounded-2xl shadow-lg transition-transform active:scale-95"
        >
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3"><polyline points="20 6 9 17 4 12"></polyline></svg>
          <span class="text-sm font-black uppercase tracking-widest">Finalizar y Firmar</span>
        </button>
      </Transition>

      <!-- Modal de Resumen y Firma -->
      <Teleport to="body">
        <Transition enter-active-class="transition-opacity duration-300" enter-from-class="opacity-0" enter-to-class="opacity-100" leave-active-class="transition-opacity duration-200" leave-from-class="opacity-100" leave-to-class="opacity-0">
          <div v-if="showModalResumen" class="fixed inset-0 z-[100] flex items-center justify-center p-4">
            <div class="absolute inset-0 bg-slate-900/60 backdrop-blur-md" @click="!isSubmittingFirm && !submitSuccess ? showModalResumen = false : null"></div>
            
            <Transition appear enter-active-class="transition-all duration-500 cubic-bezier(0.34, 1.56, 0.64, 1)" enter-from-class="opacity-0 scale-90 translate-y-10" enter-to-class="opacity-100 scale-100 translate-y-0">
              <div class="relative bg-white w-full max-w-2xl rounded-[2.5rem] shadow-2xl overflow-hidden flex flex-col max-h-[90vh]">
                
                <!-- Header -->
                <div class="px-8 pt-8 pb-6 bg-slate-50 border-b border-slate-100 flex items-center justify-between">
                  <div class="flex items-center gap-4">
                    <div class="w-12 h-12 bg-slate-900 rounded-2xl flex items-center justify-center text-white shadow-lg">
                      <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M16 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"></path><circle cx="12" cy="7" r="4"></circle></svg>
                    </div>
                    <div>
                      <h2 class="text-xl font-black text-slate-900 tracking-tight">Resumen de Novedades</h2>
                      <p class="text-xs font-bold text-slate-400 uppercase tracking-widest">Carrito {{ carritoActivo?.numero_carrito }} • {{ itemsConDiscrepancia.length }} Hallazgos</p>
                    </div>
                  </div>
                  <button @click="showModalResumen = false" class="text-slate-400 hover:text-red-500 transition-colors" :disabled="isSubmittingFirm">
                    <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><line x1="18" y1="6" x2="6" y2="18"></line><line x1="6" y1="6" x2="18" y2="18"></line></svg>
                  </button>
                </div>

                <!-- Overlay de Carga u Éxito dentro del Modal -->
                <div v-if="isSubmittingFirm || submitSuccess" class="absolute inset-x-0 bottom-0 z-50 bg-white/95 flex flex-col items-center justify-center p-8 text-center" style="top: 80px;">
                  <template v-if="isSubmittingFirm">
                    <div class="w-16 h-16 border-[5px] border-green-100 border-t-[#4cc253] rounded-full animate-spin mb-6"></div>
                    <h3 class="text-xl font-black text-slate-800 tracking-tight uppercase">Guardando Inventario</h3>
                    <p class="text-[0.75rem] font-bold text-slate-500 max-w-xs mt-2">Estamos procesando tu firma y registrando los ítems en la base de datos...</p>
                  </template>
                  <template v-else-if="submitSuccess">
                    <div class="w-20 h-20 bg-[#4cc253]/15 rounded-full flex items-center justify-center text-[#4cc253] mb-6 animate-[badge-pulse_2s_ease-in-out_infinite]">
                      <svg width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><polyline points="20 6 9 17 4 12"></polyline></svg>
                    </div>
                    <h3 class="text-2xl font-black text-slate-800 tracking-tight uppercase">¡Éxito!</h3>
                    <p class="text-[0.75rem] font-bold text-slate-500 max-w-xs mt-2">El inventario ha sido completado y firmado correctamente.</p>
                  </template>
                </div>

                <!-- Contenido -->
                <div class="flex-1 overflow-y-auto p-8 flex flex-col gap-8">
                  
                  <!-- Tabla de Hallazgos -->
                  <div v-if="itemsConDiscrepancia.length > 0" class="flex flex-col gap-3">
                    <label class="text-[0.65rem] font-black text-slate-400 uppercase tracking-[0.2em] ml-1">Detalle de Discrepancias</label>
                    <div class="bg-slate-50 rounded-3xl border border-slate-100 overflow-hidden text-center">
                      <table class="w-full text-sm border-collapse">
                        <thead>
                          <tr class="bg-slate-100/50 text-[0.6rem] font-black text-slate-500 uppercase tracking-widest border-b border-slate-100">
                            <th class="px-4 py-3 text-left">Herramienta</th>
                            <th class="px-4 py-3">Sist.</th>
                            <th class="px-4 py-3">Fís.</th>
                            <th class="px-4 py-3">Novedad</th>
                          </tr>
                        </thead>
                        <tbody class="divide-y divide-slate-100">
                          <tr v-for="h in itemsConDiscrepancia" :key="h.id" class="hover:bg-white transition-colors">
                            <td class="px-4 py-4 text-left">
                              <p class="font-extrabold text-slate-800 leading-tight">{{ h.nombre }}</p>
                              <p class="text-[0.6rem] font-bold text-slate-400 mt-1 uppercase">{{ h.codigo }}</p>
                            </td>
                            <td class="px-4 py-4 font-black text-slate-400 text-xs">{{ h.cantidadOriginal }}</td>
                            <td class="px-4 py-4 font-black text-slate-800 text-xs">{{ h.cantidad }}</td>
                            <td class="px-4 py-4 text-center flex flex-col items-center gap-1">
                              <span v-if="h.novedad !== 'Sin novedad'" :class="['text-[0.55rem] font-black px-2 py-0.5 rounded-full', h.novedad === 'Dañada' ? 'bg-orange-100 text-orange-600' : 'bg-red-100 text-red-600']">
                                {{ h.novedad.toUpperCase() }}
                              </span>
                              <span v-if="h.marca !== h.marcaOriginal" class="text-[0.55rem] font-black px-2 py-0.5 rounded-full bg-[#4cc253]/15 text-[#4cc253]">
                                MARCA MODIF.
                              </span>
                              <span v-if="h.observacion && h.observacion.trim().length > 0" class="text-[0.55rem] font-black px-2 py-0.5 rounded-full bg-slate-100 text-slate-500">
                                CON OBS.
                              </span>
                            </td>
                          </tr>
                        </tbody>
                      </table>
                    </div>
                  </div>
                  <div v-else class="flex flex-col items-center py-8 bg-[#4cc253]/5 rounded-3xl border border-[#4cc253]/20 text-center mx-2">
                    <div class="w-16 h-16 bg-[#4cc253]/15 rounded-full flex items-center justify-center text-[#4cc253] mb-3">
                       <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><polyline points="20 6 9 17 4 12"></polyline></svg>
                    </div>
                    <p class="text-sm font-black text-[#4cc253] uppercase tracking-widest">¡Auditoría Perfecta!</p>
                    <p class="text-[0.65rem] font-bold text-slate-500 tracking-wide mt-1">Todos los ítems coinciden con el sistema.</p>
                  </div>

                  <!-- Panel de Firma -->
                  <div class="flex flex-col gap-3">
                    <div class="flex items-center justify-between ml-1">
                      <label class="text-[0.65rem] font-black text-slate-400 uppercase tracking-[0.2em]">Firma Digital Obligatoria</label>
                      <button @click="clearSignature" class="text-[0.6rem] font-black text-red-400 hover:text-red-600 uppercase tracking-widest transition-colors flex items-center gap-1">
                         <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3"><path d="M3 6h18m-2 0v14c0 1-1 2-2 2H7c-1 0-2-1-2-2V6m3 0V4c0-1 1-2 2-2h4c1 0 2 1 2 2v2"></path></svg>
                         Limpiar
                      </button>
                    </div>
                    <div class="relative bg-slate-50 border-2 border-dashed border-slate-200 rounded-[2rem] h-48 overflow-hidden group hover:border-slate-300 transition-colors">
                      <canvas 
                        ref="canvasRef" 
                        width="608" 
                        height="192" 
                        class="absolute inset-0 w-full h-full cursor-crosshair touch-none"
                        @mousedown="startDrawing"
                        @mousemove="draw"
                        @mouseup="stopDrawing"
                        @mouseleave="stopDrawing"
                        @touchstart.prevent="startDrawing"
                        @touchmove.prevent="draw"
                        @touchend.prevent="stopDrawing"
                      ></canvas>
                      <div v-if="!isDrawing" class="absolute inset-0 flex items-center justify-center pointer-events-none opacity-20 group-hover:opacity-40 transition-opacity">
                        <svg width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M12 20h9"></path><path d="M16.5 3.5a2.121 2.121 0 0 1 3 3L7 19l-4 1 1-4L16.5 3.5z"></path></svg>
                      </div>
                    </div>
                  </div>
                </div>

                <!-- Footer -->
                <div class="px-8 py-6 bg-slate-50 border-t border-slate-100 flex gap-4">
                  <button @click="showModalResumen = false" :disabled="isSubmittingFirm" class="w-1/3 py-4 text-xs font-black text-slate-500 uppercase tracking-widest hover:bg-slate-200 rounded-2xl transition-colors disabled:opacity-50">Cancelar</button>
                  <button 
                    @click="confirmarYEnviar"
                    :disabled="isSubmittingFirm || submitSuccess"
                    class="w-2/3 py-4 bg-[#4cc253] hover:bg-[#43ad49] text-white rounded-2xl text-[0.7rem] font-black uppercase tracking-[0.15em] transition-transform active:scale-95 flex items-center justify-center gap-2 disabled:opacity-50 disabled:active:scale-100"
                  >
                    {{ isSubmittingFirm ? 'Enviando...' : (submitSuccess ? 'Finalizado' : 'Confirmar e Inventariar') }}
                    <svg v-if="!isSubmittingFirm && !submitSuccess" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3"><polyline points="20 6 9 17 4 12"></polyline></svg>
                    <svg v-if="submitSuccess" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3"><polyline points="20 6 9 17 4 12"></polyline></svg>
                  </button>
                </div>

              </div>
            </Transition>
          </div>
        </Transition>
      </Teleport>

      <!-- Modal para cambiar Marca -->
      <Teleport to="body">
        <Transition enter-active-class="transition-opacity duration-300" enter-from-class="opacity-0" enter-to-class="opacity-100" leave-active-class="transition-opacity duration-200" leave-from-class="opacity-100" leave-to-class="opacity-0">
          <div v-if="modalMarcaAbierto" class="fixed inset-0 z-[110] flex items-center justify-center p-4">
            <div class="absolute inset-0 bg-slate-900/60 backdrop-blur-md" @click="modalMarcaAbierto = false"></div>
            
            <Transition appear enter-active-class="transition-all duration-300 ease-out" enter-from-class="opacity-0 scale-95 translate-y-4" enter-to-class="opacity-100 scale-100 translate-y-0">
              <div v-if="modalMarcaAbierto" class="relative bg-white w-full max-w-md rounded-3xl shadow-2xl p-6 flex flex-col gap-4 max-h-[80vh]">
                <div class="flex justify-between items-center">
                  <h3 class="text-xl font-black text-slate-800">Seleccionar Marca</h3>
                  <button @click="modalMarcaAbierto = false" class="text-slate-400 hover:text-red-500 transition-colors">
                    <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><line x1="18" y1="6" x2="6" y2="18"></line><line x1="6" y1="6" x2="18" y2="18"></line></svg>
                  </button>
                </div>

                <div v-if="herramientaParaMarca" class="bg-slate-50 p-3 rounded-2xl border border-slate-100">
                  <p class="text-[0.65rem] font-bold text-slate-400 uppercase tracking-widest">Herramienta Actual</p>
                  <p class="text-sm font-extrabold text-slate-700 leading-tight">{{ herramientaParaMarca.nombre }}</p>
                </div>

                <div class="relative">
                  <div class="absolute inset-y-0 left-3 flex items-center pointer-events-none">
                     <svg class="text-slate-400" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><circle cx="11" cy="11" r="8"></circle><line x1="21" y1="21" x2="16.65" y2="16.65"></line></svg>
                  </div>
                  <input v-model="busquedaMarcaModal" type="text" placeholder="Buscar nueva marca..." class="w-full py-3 pl-10 pr-4 bg-white border-2 border-slate-200 rounded-xl text-sm font-semibold focus:outline-none focus:border-green-400" />
                </div>

                <div class="flex-1 overflow-y-auto pr-2" style="min-height: 250px;">
                  <div v-if="isLoadingMarcas" class="flex flex-col gap-2 p-1">
                    <div v-for="i in 5" :key="i" class="w-full flex items-center justify-between px-4 py-4 bg-slate-50 border border-slate-100 rounded-xl animate-pulse">
                      <div class="h-3 bg-slate-200 rounded-full w-1/2"></div>
                      <div class="w-4 h-4 bg-slate-200 rounded-full"></div>
                    </div>
                  </div>
                  <ul v-else class="flex flex-col gap-2">
                     <li v-for="m in marcasModalFiltradas" :key="m" @click="seleccionarNuevaMarca(m)" class="px-4 py-3 rounded-xl bg-white border border-slate-100 hover:bg-green-50 hover:border-green-300 font-bold text-slate-700 text-sm cursor-pointer transition-colors flex justify-between items-center group">
                       <span v-html="resaltarTexto(m, busquedaMarcaModal)"></span>
                       <svg class="opacity-0 group-hover:opacity-100 text-green-500 transition-opacity" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3"><polyline points="20 6 9 17 4 12"></polyline></svg>
                     </li>
                     <li v-if="marcasModalFiltradas.length === 0" class="text-center p-4 text-xs font-bold text-slate-400">
                        No se encontraron marcas
                     </li>
                  </ul>
                </div>
              </div>
            </Transition>
          </div>
        </Transition>
      </Teleport>

      <!-- Estado Auditado Completamente -->
      <div v-if="carritoYaFirmado" class="flex flex-col items-center gap-5 py-24 text-center">
        <div class="w-24 h-24 rounded-full bg-[#4cc253]/10 flex items-center justify-center mb-2 animate-[badge-pulse_3s_ease-in-out_infinite]">
          <svg width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="#4cc253" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
            <polyline points="20 6 9 17 4 12"></polyline>
          </svg>
        </div>
        <h3 class="text-3xl font-black text-slate-800 tracking-tight uppercase">Auditoría Finalizada</h3>
        <p class="text-[0.8rem] font-bold text-slate-500 max-w-sm">Este carrito ya fue inventariado y firmado exitosamente en el sistema.</p>
        <button @click="dropdownAbierto = true; window.scrollTo({top: 0, behavior: 'smooth'})" class="mt-4 px-8 py-4 bg-slate-100 hover:bg-slate-200 text-slate-600 rounded-2xl font-black text-[0.65rem] uppercase tracking-[0.15em] transition-colors">
          Seleccionar otro carrito
        </button>
      </div>

      <!-- Estado vacío (Sin Resultados de Filtro) -->
      <div v-else-if="!isLoadingItems && herramientasFiltradas.length === 0" class="flex flex-col items-center gap-4 py-20 text-center">
        <div class="w-20 h-20 rounded-3xl bg-slate-100 flex items-center justify-center">
          <svg width="36" height="36" viewBox="0 0 24 24" fill="none" stroke="#94a3b8" stroke-width="1.5"><circle cx="11" cy="11" r="8"></circle><line x1="21" y1="21" x2="16.65" y2="16.65"></line></svg>
        </div>
        <h3 class="text-lg font-black text-slate-700">Sin resultados</h3>
        <p class="text-sm font-medium text-slate-400 max-w-xs">Intenta con otros filtros o limpia la búsqueda con el botón de reset.</p>
      </div>
    </main>
  </div>
</template>

<style>
@keyframes card-fly-in {
  from { 
    opacity: 0; 
    transform: translateY(30px) scale(0.95); 
    filter: blur(10px);
  }
  to { 
    opacity: 1; 
    transform: translateY(0) scale(1); 
    filter: blur(0);
  }
}
.card-animated { 
  animation: card-fly-in 0.7s cubic-bezier(0.34, 1.56, 0.64, 1) both; 
}

@keyframes badge-pulse {
  0%, 100% { box-shadow: 0 0 0 0 rgba(34, 197, 94, 0); }
  50%       { box-shadow: 0 0 0 5px rgba(34, 197, 94, 0.15); }
}
.badge-pulse { animation: badge-pulse 3s ease-in-out infinite; }

.save-btn::after {
  content: '';
  position: absolute;
  top: 0; left: -100%;
  width: 60%; height: 100%;
  background: linear-gradient(90deg, transparent, rgba(255,255,255,0.28), transparent);
  transition: left 0.5s ease;
  pointer-events: none;
}
.save-btn:hover::after { left: 160%; }
.save-btn:hover { filter: brightness(0.93); transform: translateY(-2px); box-shadow: 0 8px 22px rgba(34,197,94,0.4) !important; }
.save-btn:active { transform: translateY(0) scale(0.98); }

.pv-field .p-textarea {
  border-radius: 12px !important;
  border: 1.5px solid #e2e8f0 !important;
  font-family: inherit !important;
  font-size: 0.88rem !important;
  padding: 0.7rem 0.9rem !important;
  resize: none !important;
}
.pv-field .p-textarea:focus {
  border-color: #4ade80 !important;
  box-shadow: 0 0 0 3px rgba(74,222,128,0.15) !important;
  outline: none !important;
}
.pv-field .p-textarea::placeholder { color: #cbd5e1 !important; }

.staggered-fade-enter-active { transition: all 0.5s cubic-bezier(0.34, 1.56, 0.64, 1); }
.staggered-fade-leave-active { transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1); }
.staggered-fade-enter-from, .staggered-fade-leave-to { opacity: 0; transform: translateY(20px) scale(0.95); }
.staggered-fade-move { transition: transform 0.5s cubic-bezier(0.34, 1.56, 0.64, 1); }

.fade-slide-enter-active { transition: all 0.3s cubic-bezier(0.34, 1.56, 0.64, 1); }
.fade-slide-leave-active { transition: all 0.2s ease-in; }
.fade-slide-enter-from, .fade-slide-leave-to { opacity: 0; transform: translateY(-10px) scale(0.95); }
</style>
