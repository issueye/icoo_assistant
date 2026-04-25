<template>
  <label class="block">
    <span class="mb-2 block text-sm font-semibold text-slate-200">{{ label }}</span>
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
