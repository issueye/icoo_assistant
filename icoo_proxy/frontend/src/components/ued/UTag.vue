<template>
  <component
    :is="as"
    class="ued-tag"
    :class="[`ued-tag--${normalizedVariant}`, `ued-tag--${normalizedSize}`, { 'ued-tag--code': code, 'ued-tag--dot': dot }]"
  >
    <span v-if="dot" class="ued-tag__dot" aria-hidden="true"></span>
    <span class="ued-tag__content">
      <slot />
    </span>
  </component>
</template>

<script setup>
import { computed } from "vue";

const props = defineProps({
  variant: {
    type: String,
    default: "neutral",
  },
  size: {
    type: String,
    default: "sm",
  },
  code: {
    type: Boolean,
    default: false,
  },
  dot: {
    type: Boolean,
    default: false,
  },
  as: {
    type: String,
    default: "span",
  },
});

const normalizedVariant = computed(() => {
  const value = String(props.variant || "neutral").toLowerCase();
  if (value === "danger") {
    return "error";
  }
  if (value === "sucess") {
    return "success";
  }
  return value;
});

const normalizedSize = computed(() => {
  const value = String(props.size || "sm").toLowerCase();
  return ["xs", "sm", "md", "lg"].includes(value) ? value : "sm";
});
</script>
