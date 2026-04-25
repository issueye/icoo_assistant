<template>
  <article class="stat-card">
    <p class="stat-label">{{ label }}</p>
    <p class="stat-value">{{ value }}</p>
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
