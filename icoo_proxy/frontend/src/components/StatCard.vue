<template>
  <article class="rounded-3xl border border-white/10 bg-ink-900/80 p-4">
    <p class="text-xs uppercase tracking-[0.2em] text-slate-400">{{ label }}</p>
    <p class="mt-3 break-all text-lg font-semibold">{{ value }}</p>
  </article>
</template>

<script setup>
defineProps({
  label: {
    type: String,
    required: true,
  },
  value: {
    type: String,
    required: true,
  },
});
</script>
