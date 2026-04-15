<template>
  <div class="hero-wrapper relative w-full h-full flex flex-col justify-between overflow-hidden">
    <div class="absolute inset-0" style="background: linear-gradient(to bottom, rgba(0,0,0,0.55) 0%, rgba(0,0,0,0.2) 40%, rgba(0,0,0,0.85) 100%)" />

    <div class="relative z-10 flex items-center gap-2.5">
      <img src="/favicon.ico" alt="SAO6 Logo" class="logo-blanco" width="32" height="32" />
      <span class="font-bold text-xl text-white tracking-widest">SAO6</span>
    </div>

    <div class="relative z-10">
      <h1 class="font-black text-white leading-tight mb-5" style="font-size: clamp(2.8rem, 5vw, 3.8rem)">
        ¡Medellín<br />Conduce el<br />
        <span class="text-brand">Cambio!</span>
      </h1>
      <p class="text-base text-white/80 leading-relaxed max-w-sm">
        Nuestros buses alimentadores no solo llevan personas: impulsan los sueños,
        metas e historias que construyen ciudad.
      </p>
    </div>

    <div class="absolute bottom-8 right-8 z-10 max-md:hidden">
      <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="white" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <path d="M12 5v14M5 12l7 7 7-7"/>
      </svg>
    </div>
  </div>
</template>

<style scoped>
.hero-wrapper {
  background-image: url('/backgroundAutoSao.webp');
  background-size: cover;
  background-position: center;
  padding: 2.5rem 1.5rem;
  /* Espaciado extra abajo para que el formulario no tape el texto en móvil */
  padding-bottom: 4rem; 
}

@media (min-width: 768px) {
  .hero-wrapper {
    background-position: 70% center;
    padding: 3rem 4rem;
    padding-bottom: 3rem;
  }
}

.logo-blanco {
  filter: brightness(0) invert(1);
  object-fit: contain;
}
</style>
