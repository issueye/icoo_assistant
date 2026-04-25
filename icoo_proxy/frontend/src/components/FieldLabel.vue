<template>
  <label class="block">
    <span class="mb-1.5 block text-sm font-medium text-slate-700">{{ label }}</span>
    <slot />
  </label>
</template>

<script setup>
defineProps({
  label: {
    type: String,
    required: true,
  },
});
</script>
