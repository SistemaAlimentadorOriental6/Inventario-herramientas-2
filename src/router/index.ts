import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../domains/auth/store/useAuthStore'

export const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      redirect: '/login'
    },
    {
      path: '/login',
      name: 'login',
      component: () => import('../domains/auth/views/LoginView.vue'),
      meta: { requiresAuth: false }
    },
    {
      path: '/dashboard/stock',
      name: 'stockDashboard',
      component: () => import('../domains/stock/views/StockDashboard.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/dashboard/prestamo',
      name: 'prestamoDashboard',
      component: () => import('../domains/prestamo/views/PrestamoDashboard.vue'),
      meta: { requiresAuth: true, role: 'admin' }
    },
    {
      path: '/dashboard/admin',
      name: 'adminDashboard',
      component: () => import('../domains/admin/views/AdminDashboard.vue'),
      meta: { requiresAuth: true, role: 'admin' }
    }
  ]
})

router.beforeEach((to, from, next) => {
  const authStore = useAuthStore()
  const estaAutenticado = !!authStore.token
  const rol = authStore.usuario?.rol

  // Redirigir si requiere auth y no está logueado
  if (to.meta.requiresAuth && !estaAutenticado) {
    return next('/login')
  }

  // Redirigir si ya está logueado e intenta ir al login
  if (to.name === 'login' && estaAutenticado) {
    if (rol === 'admin') return next('/dashboard/admin')
    return next('/dashboard/stock')
  }

  // Acceso especial a prestamo: solo cedula 1083462461 (campo empleado)
  const cedulaPrestamo = authStore.usuario?.empleado
  if (to.path === '/dashboard/prestamo') {
    if (cedulaPrestamo !== '1083462461') {
      return next('/dashboard/admin')
    }
    return next()
  }

  // Admin: solo puede estar en /dashboard/admin
  if (estaAutenticado && rol === 'admin') {
    if (to.path.startsWith('/dashboard') && to.path !== '/dashboard/admin') {
      return next('/dashboard/admin')
    }
  }

  // Operario: solo puede estar en /dashboard/stock
  if (estaAutenticado && rol === 'operario') {
    if (to.path.startsWith('/dashboard') && to.path !== '/dashboard/stock') {
      return next('/dashboard/stock')
    }
  }

  next()
})

