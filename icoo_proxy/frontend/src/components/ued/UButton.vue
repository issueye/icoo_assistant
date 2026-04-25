<template>
  <button
    :type="nativeType"
    class="ued-button"
    :class="[
      `ued-button--${normalizedVariant}`,
      `ued-button--${normalizedSize}`,
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
import { computed } from "vue";

defineEmits(["click"]);

const props = defineProps({
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

const normalizedVariant = computed(() => {
  const value = String(props.variant || "primary").toLowerCase();
  if (value === "danger") {
    return "error";
  }
  if (value === "sucess") {
    return "success";
  }
  return value;
});

const normalizedSize = computed(() => {
  const value = String(props.size || "md").toLowerCase();
  return ["xs", "sm", "md", "lg"].includes(value) ? value : "md";
});
</script>
