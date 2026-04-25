<template>
  <div v-if="items.length === 0" class="rounded-3xl border border-white/10 bg-black/20 px-4 py-8 text-center text-sm text-slate-400">
    {{ emptyText }}
  </div>
  <div v-else class="space-y-3">
    <article v-for="item in items" :key="item.name" class="flex flex-wrap items-center justify-between gap-3 rounded-3xl border border-white/10 bg-ink-900/70 p-4">
      <div>
        <p class="text-sm font-semibold">{{ item.name }}</p>
        <p class="mt-1 text-xs text-slate-400">{{ item.upstream }}</p>
      </div>
      <code class="rounded-full bg-black/20 px-3 py-2 font-mono text-xs">{{ item.model }}</code>
    </article>
  </div>
</template>

<script setup>
defineProps({
  items: {
    type: Array,
    default: () => [],
  },
  emptyText: {
    type: String,
    default: "No items.",
  },
});
</script>
