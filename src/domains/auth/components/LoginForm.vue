<script setup lang="ts">
import { ref, computed } from 'vue'
import InputText from 'primevue/inputtext'
import Password from 'primevue/password'
import Button from 'primevue/button'
import { useAuthStore } from '../store/useAuthStore'
import type { CredencialesLogin, CredencialesLoginCedula } from '../types/auth.types'
import { useRouter } from 'vue-router'

const authStore = useAuthStore()
const router = useRouter()

// Formulario unificado
const identificador = ref('')
const contrasena = ref('')

// Detecta si es email (admin/operario) o cédula (empleado) o usuario especial
const esEmail = computed(() => identificador.value.includes('@') || identificador.value.trim() === 'inventario_herramientas')
const esCedula = computed(() => /^\d+$/.test(identificador.value.trim()))

const puedeEnviar = computed(() => {
  const id = identificador.value.trim()
  const pass = contrasena.value.trim()
  
  if (esEmail.value) {
    // Admin/Operario: email + contraseña
    return id !== '' && pass !== ''
  }
  if (esCedula.value) {
    // Empleado: cédula + confirmación de cédula (deben coincidir)
    return id !== '' && pass !== '' && id === pass
  }
  return false
})

const tipoUsuarioLabel = computed(() => {
  if (esEmail.value) return 'Admin / Operario'
  if (esCedula.value) return 'Empleado'
  return ''
})

async function onSubmit() {
  if (esEmail.value) {
    // Login como admin/operario
    await authStore.login({ email: identificador.value.trim(), contrasena: contrasena.value })
  } else {
    // Login como empleado por cédula (la contraseña debe ser la misma cédula)
    await authStore.loginPorCedula({ cedula: identificador.value.trim() })
  }

  if (!authStore.error) {
    // Solo el usuario especial de inventario entra a /dashboard/prestamo
    const user = authStore.usuario
    const idIngresado = identificador.value.trim().toLowerCase()
    const emailReal = user?.email?.toLowerCase()
    
    // Los usuarios con rol 'inventario' van exclusivamente a /dashboard/prestamo
    const esUsuarioPrestamo = user?.rol === 'inventario'

    router.push(esUsuarioPrestamo ? '/dashboard/prestamo' : '/dashboard/stock')
  }
}
</script>

<template>
  <div class="form-wrapper flex flex-col justify-center mx-auto" style="width: 100%; max-width: 460px; padding: 2.5rem 1.5rem; box-sizing: border-box;">
    <div class="form-header mb-10 text-center flex flex-col items-center">
      <h2 class="font-extrabold text-4xl text-gray-900 mb-4 flex items-center justify-center gap-3">
        Bienvenido
        <span class="wave-emoji">👋</span>
      </h2>
      <p class="text-base sm:text-lg text-slate-500 mt-2">Ingresa tus credenciales para continuar.</p>
    </div>

    <!-- Formulario unificado -->
    <form class="flex flex-col form-body w-full" @submit.prevent="onSubmit" novalidate>
      <!-- Campo 1: Usuario / Email / Cédula -->
      <div class="field-group form-field text-left flex flex-col">
        <label class="field-label font-semibold text-gray-600 block" for="identificador">
          {{ esCedula ? 'CÉDULA' : 'EMAIL' }}
        </label>
        <InputText
          id="identificador"
          v-model="identificador"
          :placeholder="esCedula ? 'Ej: 1037618963' : 'Ej: usuario@sao6.com.co'"
          :type="esCedula ? 'text' : 'email'"
          :inputmode="esCedula ? 'numeric' : 'email'"
          autocomplete="username"
          fluid
          class="pv-input w-full"
        />
      </div>

      <!-- Campo 2: Contraseña (admin/operario) o Cédula (empleado) -->
      <div class="field-group form-field text-left flex flex-col">
        <label class="field-label flex justify-between font-semibold text-gray-600 block" for="credencial">
          <span>{{ esCedula ? 'CONFIRMAR CÉDULA' : 'CONTRASEÑA' }}</span>
        </label>
        <Password
          v-model="contrasena"
          inputId="credencial"
          :placeholder="esCedula ? 'Repite tu cédula' : '••••••••'"
          :feedback="false"
          :toggleMask="!esCedula"
          :autocomplete="esCedula ? 'off' : 'current-password'"
          fluid
          class="pv-input w-full"
        />
      </div>

      <Transition name="error-bounce">
        <div v-if="authStore.error" class="error-alert">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" class="error-icon" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
            <circle cx="12" cy="12" r="10"></circle>
            <line x1="12" y1="8" x2="12" y2="12"></line>
            <line x1="12" y1="16" x2="12.01" y2="16"></line>
          </svg>
          <span class="error-text">{{ authStore.error }}</span>
        </div>
      </Transition>

      <Button
        type="submit"
        :label="esCedula ? 'Ver Mis Herramientas' : 'Iniciar Sesión'"
        :loading="authStore.cargando"
        :disabled="!puedeEnviar"
        fluid
        class="pv-btn w-full mt-4"
      />
    </form>
  </div>
</template>

<style>

.form-wrapper {
  animation: formFadeUp 0.5s cubic-bezier(0.22, 1, 0.36, 1) both;
}

.form-header {
  margin-bottom: 2.5rem; /* Separación forzada del subtítulo con los campos */
  animation: formFadeUp 0.55s 0.05s cubic-bezier(0.22, 1, 0.36, 1) both;
}

.form-body {
  gap: 1.5rem;
}

.form-field:nth-child(1) { animation: formFadeUp 0.55s 0.1s cubic-bezier(0.22, 1, 0.36, 1) both; }
.form-field:nth-child(2) { animation: formFadeUp 0.55s 0.18s cubic-bezier(0.22, 1, 0.36, 1) both; }
.form-field:nth-child(3) { animation: formFadeUp 0.55s 0.26s cubic-bezier(0.22, 1, 0.36, 1) both; }

@keyframes formFadeUp {
  from { opacity: 0; transform: translateY(18px); }
  to   { opacity: 1; transform: translateY(0); }
}

/* ── Emoji de saludo animado ── */
.wave-emoji {
  display: inline-block;
  animation: wave 2s ease-in-out infinite;
  transform-origin: 70% 70%;
}

@keyframes wave {
  0%, 100% { transform: rotate(0deg); }
  10% { transform: rotate(14deg); }
  20% { transform: rotate(-8deg); }
  30% { transform: rotate(14deg); }
  40% { transform: rotate(-4deg); }
  50% { transform: rotate(10deg); }
  60% { transform: rotate(0deg); }
}

/* ── Grupo de campo ── */
.field-group {
  display: flex;
  flex-direction: column;
}

.field-label {
  font-size: 0.75rem;
  font-weight: 600;
  letter-spacing: 0.05em;
  color: #64748b;
  text-transform: uppercase;
  transition: color 0.2s;
  display: block;
  margin-bottom: 0.65rem; /* Margen inferior forzado para separar del input */
}

.field-group:focus-within .field-label {
  color: #4cc253;
}

/* ── InputText ── */
.pv-input .p-inputtext,
.pv-input.p-inputtext {
  font-family: var(--font-sans);
  font-size: 0.95rem;
  border-radius: 16px;
  border: none;
  background: #f1f5f9;
  box-shadow: none;
  padding: 1rem 1.25rem;
  width: 100%;
  transition: background 0.2s, box-shadow 0.2s;
  color: #334155;
}

.pv-input .p-inputtext::placeholder,
.pv-input.p-inputtext::placeholder {
  color: #94a3b8;
}

.pv-input .p-inputtext:hover,
.pv-input.p-inputtext:hover {
  background: #e2e8f0;
}

.pv-input .p-inputtext:focus,
.pv-input.p-inputtext:focus {
  background: #fff !important;
  box-shadow: 0 0 0 2px #4cc253 !important;
  outline: none;
}

/* ── Password ── */
.pv-input.p-password {
  width: 100%;
}

.pv-input .p-password-input {
  font-family: var(--font-sans);
  font-size: 0.95rem;
  border-radius: 16px;
  border: none;
  background: #f1f5f9;
  box-shadow: none;
  padding: 1rem 1.25rem;
  width: 100%;
  transition: background 0.2s, box-shadow 0.2s;
  color: #334155;
}

.pv-input .p-password-input::placeholder {
  color: #94a3b8;
}

.pv-input .p-password-input:hover {
  background: #e2e8f0;
}

.pv-input .p-password-input:focus {
  background: #fff !important;
  box-shadow: 0 0 0 2px #4cc253 !important;
  outline: none;
}

/* Icono del ojo en password */
.pv-input .p-password-toggle {
  color: #94a3b8 !important;
  margin-right: 0.5rem;
}

.pv-input .p-password-toggle:hover {
  color: #64748b !important;
}

/* ── Botón ── */
.pv-btn.p-button {
  min-height: 56px;
  font-family: var(--font-sans) !important;
  font-weight: 600 !important;
  font-size: 1.05rem !important;
  letter-spacing: 0 !important;
  border-radius: 16px !important;
  background: linear-gradient(135deg, #4ade80 0%, #22c55e 100%) !important;
  border: none !important;
  box-shadow: 0 8px 20px rgba(34, 197, 94, 0.35) !important;
  transition: all 0.2s ease !important;
  animation: formFadeUp 0.55s 0.32s cubic-bezier(0.22, 1, 0.36, 1) both;
  margin-top: 1rem;
}

.pv-btn.p-button:not(:disabled):hover {
  background: linear-gradient(135deg, #22c55e 0%, #16a34a 100%) !important;
  box-shadow: 0 12px 28px rgba(34, 197, 94, 0.45) !important;
  transform: translateY(-2px) !important;
}

.pv-btn.p-button:not(:disabled):active {
  transform: translateY(0) !important;
  box-shadow: 0 4px 12px rgba(34, 197, 94, 0.35) !important;
}

.pv-btn.p-button:disabled {
  background: #cbd5e1 !important;
  border: none !important;
  color: #94a3b8 !important;
  box-shadow: none !important;
  transform: none !important;
}

/* ── Transición y Diseño del mensaje de error ── */
.error-alert {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  background: linear-gradient(to right, #fef2f2, #fff1f2);
  border: 1px solid #fecdd3;
  color: #e11d48;
  border-radius: 9999px; /* Aspecto de pastilla elegante */
  padding: 0.75rem 1.25rem;
  margin-bottom: 0.5rem;
  box-shadow: 0 4px 15px rgba(225, 29, 72, 0.08); /* Sombra difuminada muy premium */
  overflow: hidden;
  max-height: 60px; /* Base height for transition */
}

.error-text {
  font-size: 0.85rem;
  font-weight: 600;
  letter-spacing: 0.01em;
}

.error-icon {
  flex-shrink: 0;
  color: #f43f5e;
}

.error-bounce-enter-active,
.error-bounce-leave-active {
  transition: all 0.4s cubic-bezier(0.25, 1, 0.5, 1);
}

.error-bounce-enter-from,
.error-bounce-leave-to {
  opacity: 0;
  transform: scale(0.95) translateY(-5px);
  max-height: 0;
  margin-bottom: 0;
  padding-top: 0;
  padding-bottom: 0;
  border-width: 0;
}

/* ── Tabs de Login ── */
.login-tabs {
  display: flex;
  gap: 0.5rem;
  background: #f1f5f9;
  padding: 0.375rem;
  border-radius: 12px;
  animation: formFadeUp 0.55s 0.08s cubic-bezier(0.22, 1, 0.36, 1) both;
}

.tab-btn {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  padding: 0.75rem 1rem;
  font-size: 0.8rem;
  font-weight: 700;
  border-radius: 10px;
  border: none;
  cursor: pointer;
  transition: all 0.2s ease;
}

.tab-inactive {
  background: transparent;
  color: #64748b;
}

.tab-inactive:hover {
  background: rgba(255, 255, 255, 0.5);
  color: #334155;
}

.tab-active {
  background: white;
  color: #22c55e;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
}
</style>
