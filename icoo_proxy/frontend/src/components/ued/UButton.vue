<template>
  <button
    :type="nativeType"
    class="ued-button"
    :class="[
      `ued-button--${variant}`,
      `ued-button--${size}`,
      {
        'ued-button--block': block,
        'is-loading': loading,
      },
    ]"
    :disabled="disabled || loading"
    @click="$emit('click', $event)"
  >
    <span v-if="loading" class="ued-button__spinner" />
    <span><slot /></span>
  </button>
</template>

<script setup>
defineEmits(["click"]);

defineProps({
  variant: {
    type: String,
    default: "primary",
  },
  size: {
    type: String,
    default: "md",
  },
  block: {
    type: Boolean,
    default: false,
  },
  loading: {
    type: Boolean,
    default: false,
  },
  disabled: {
    type: Boolean,
    default: false,
  },
  nativeType: {
    type: String,
    default: "button",
  },
});
</script>
