<template>
  <section class="content-panel">
    <div class="panel-header">
      <p v-if="eyebrow" class="panel-eyebrow">{{ eyebrow }}</p>
      <h3 class="panel-title">{{ title }}</h3>
    </div>
    <div class="panel-body">
      <slot />
    </div>
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
