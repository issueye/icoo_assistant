<template>
  <section class="rounded-[30px] border border-white/10 bg-white/5 p-5 shadow-panel backdrop-blur">
    <div class="mb-5">
      <p class="text-xs uppercase tracking-[0.24em] text-signal-sky">{{ eyebrow }}</p>
      <h3 class="mt-2 text-2xl font-bold tracking-[-0.04em]">{{ title }}</h3>
    </div>
    <slot />
  </section>
</template>

<script setup>
defineProps({
  title: {
    type: String,
    required: true,
  },
  eyebrow: {
    type: String,
    default: "",
  },
});
</script>
