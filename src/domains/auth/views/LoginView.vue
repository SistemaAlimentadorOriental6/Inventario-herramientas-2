<script setup lang="ts">
import LoginHero from '../components/LoginHero.vue'
import LoginForm from '../components/LoginForm.vue'
</script>

<template>
  <div class="flex flex-col min-h-screen bg-gray-50 md:grid md:grid-cols-2">
    <aside class="relative w-full h-[45vh] md:h-auto">
      <LoginHero />
    </aside>
    
    <main class="mobile-curve-container relative z-20 flex items-center justify-center bg-white h-full w-full">
      <LoginForm />
    </main>
  </div>
</template>

<style scoped>
.mobile-curve-container {
  border-radius: 40px 40px 0 0;
  margin-top: -3rem; /* Solapamiento de la curva con el Hero */
  padding-top: 2.5rem;
  box-shadow: 0 -8px 24px rgba(0, 0, 0, 0.1); /* Efecto visual elegante para acentuar el solapamiento */
}

@media (min-width: 768px) {
  .mobile-curve-container {
    border-radius: 0;
    margin-top: 0;
    padding-top: 0;
    box-shadow: none;
    padding: 2rem;
  }
}
</style>
