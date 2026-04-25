<template>
  <div v-if="items.length === 0" class="empty-state">
    {{ emptyText }}
  </div>
  <div v-else class="space-y-3">
    <article v-for="item in items" :key="item.name" class="table-row">
      <div>
        <p class="text-sm font-medium text-slate-900">{{ item.name }}</p>
        <p class="mt-1 text-xs text-slate-500">{{ item.upstream }}</p>
      </div>
      <code class="mono-chip">{{ item.model }}</code>
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
