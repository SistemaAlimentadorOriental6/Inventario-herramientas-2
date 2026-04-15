import { ref } from 'vue'
import { useAuthStore } from '../store/useAuthStore'
import type { CredencialesLogin } from '../types/auth.types'

export function useLogin() {
  const authStore = useAuthStore()

  const form = ref<CredencialesLogin>({
    email: '',
    contrasena: '',
  })

  const mostrarContrasena = ref(false)

  async function submitLogin() {
    if (!form.value.email || !form.value.contrasena) return
    await authStore.login(form.value)
  }

  function toggleContrasena() {
    mostrarContrasena.value = !mostrarContrasena.value
  }

  return {
    form,
    mostrarContrasena,
    cargando: authStore.$state,
    error: authStore.error,
    submitLogin,
    toggleContrasena,
  }
}
