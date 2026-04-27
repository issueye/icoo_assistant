<template>
  <label class="ued-field">
    <span class="ued-field__label">
      {{ label }}
      <span v-if="required" class="ued-field__required" aria-hidden="true">*</span>
    </span>
    <slot />
    <p v-if="hint && !error" class="ued-field__hint">{{ hint }}</p>
    <p v-if="error" class="ued-field__error">{{ error }}</p>
  </label>
</template>

<script setup>
defineProps({
  label: {
    type: String,
    required: true,
  },
  hint: {
    type: String,
    default: "",
  },
  error: {
    type: String,
    default: "",
  },
  required: {
    type: Boolean,
    default: false,
  },
});
</script>
