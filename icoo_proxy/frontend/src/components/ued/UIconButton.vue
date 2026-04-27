<template>
  <button
    :type="nativeType"
    class="ued-icon-button"
    :class="[`ued-icon-button--${normalizedVariant}`, `ued-icon-button--${normalizedSize}`, { 'is-loading': loading }]"
    :title="label"
    :aria-label="label"
    :disabled="disabled || loading"
    @click="$emit('click', $event)"
  >
    <span v-if="loading" class="ued-icon-button__spinner" aria-hidden="true"></span>
    <span v-else class="ued-icon-button__icon" aria-hidden="true">
      <svg v-if="normalizedIcon === 'edit'" viewBox="0 0 24 24">
        <path d="M12 20h9" />
        <path d="M16.5 3.5a2.1 2.1 0 0 1 3 3L7 19l-4 1 1-4Z" />
      </svg>
      <svg v-else-if="normalizedIcon === 'delete'" viewBox="0 0 24 24">
        <path d="M3 6h18" />
        <path d="M8 6V4h8v2" />
        <path d="M19 6l-1 14H6L5 6" />
        <path d="M10 11v5" />
        <path d="M14 11v5" />
      </svg>
      <svg v-else-if="normalizedIcon === 'copy'" viewBox="0 0 24 24">
        <rect x="8" y="8" width="12" height="12" rx="2" />
        <path d="M16 8V6a2 2 0 0 0-2-2H6a2 2 0 0 0-2 2v8a2 2 0 0 0 2 2h2" />
      </svg>
      <svg v-else-if="normalizedIcon === 'check'" viewBox="0 0 24 24">
        <path d="M20 6 9 17l-5-5" />
      </svg>
      <svg v-else-if="normalizedIcon === 'inspect'" viewBox="0 0 24 24">
        <path d="M22 12h-4l-3 8-6-16-3 8H2" />
      </svg>
      <svg v-else-if="normalizedIcon === 'models'" viewBox="0 0 24 24">
        <path d="M12 2 3 7l9 5 9-5Z" />
        <path d="m3 12 9 5 9-5" />
        <path d="m3 17 9 5 9-5" />
      </svg>
      <svg v-else-if="normalizedIcon === 'detail'" viewBox="0 0 24 24">
        <path d="M2 12s3.5-6 10-6 10 6 10 6-3.5 6-10 6-10-6-10-6Z" />
        <circle cx="12" cy="12" r="3" />
      </svg>
      <svg v-else viewBox="0 0 24 24">
        <circle cx="12" cy="12" r="9" />
        <path d="M12 8v4" />
        <path d="M12 16h.01" />
      </svg>
    </span>
  </button>
</template>

<script setup>
import { computed } from "vue";

defineEmits(["click"]);

const props = defineProps({
  icon: {
    type: String,
    default: "detail",
  },
  label: {
    type: String,
    required: true,
  },
  variant: {
    type: String,
    default: "secondary",
  },
  size: {
    type: String,
    default: "sm",
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

const normalizedIcon = computed(() => String(props.icon || "detail").toLowerCase());

const normalizedVariant = computed(() => {
  const value = String(props.variant || "secondary").toLowerCase();
  if (value === "danger") {
    return "error";
  }
  return ["secondary", "primary", "info", "success", "warning", "error", "ghost"].includes(value) ? value : "secondary";
});

const normalizedSize = computed(() => {
  const value = String(props.size || "sm").toLowerCase();
  return ["xs", "sm", "md"].includes(value) ? value : "sm";
});
</script>

<style scoped>
.ued-icon-button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex: 0 0 auto;
  border: 1px solid transparent;
  border-radius: var(--ued-radius-sm);
  background: transparent;
  color: #526174;
  transition: background-color 0.16s ease, border-color 0.16s ease, color 0.16s ease, transform 0.16s ease;
}

.ued-icon-button:hover:not(:disabled) {
  transform: translateY(-1px);
}

.ued-icon-button:active:not(:disabled) {
  transform: translateY(0);
}

.ued-icon-button:disabled {
  opacity: 0.55;
}

.ued-icon-button--xs {
  width: 26px;
  height: 26px;
}

.ued-icon-button--sm {
  width: 30px;
  height: 30px;
}

.ued-icon-button--md {
  width: 34px;
  height: 34px;
}

.ued-icon-button__icon,
.ued-icon-button__spinner {
  width: 16px;
  height: 16px;
}

.ued-icon-button__icon svg {
  width: 16px;
  height: 16px;
  fill: none;
  stroke: currentColor;
  stroke-width: 2;
  stroke-linecap: round;
  stroke-linejoin: round;
}

.ued-icon-button__spinner {
  border: 2px solid currentColor;
  border-right-color: transparent;
  border-radius: 999px;
  animation: ued-icon-spin 0.8s linear infinite;
}

.ued-icon-button--secondary,
.ued-icon-button--ghost {
  color: #526174;
}

.ued-icon-button--secondary:hover:not(:disabled),
.ued-icon-button--ghost:hover:not(:disabled) {
  border-color: #cfd8e6;
  background: #f8fafc;
  color: #172033;
}

.ued-icon-button--primary {
  color: var(--ued-color-primary);
}

.ued-icon-button--primary:hover:not(:disabled) {
  border-color: #b9c8ff;
  background: var(--ued-color-primary-soft);
}

.ued-icon-button--info {
  color: var(--ued-color-info);
}

.ued-icon-button--info:hover:not(:disabled) {
  border-color: #9fe0eb;
  background: var(--ued-color-info-soft);
}

.ued-icon-button--success {
  color: var(--ued-color-success);
}

.ued-icon-button--success:hover:not(:disabled) {
  border-color: #a9dec5;
  background: var(--ued-color-success-soft);
}

.ued-icon-button--warning {
  color: var(--ued-color-warning);
}

.ued-icon-button--warning:hover:not(:disabled) {
  border-color: #f0cb85;
  background: var(--ued-color-warning-soft);
}

.ued-icon-button--error {
  color: var(--ued-color-destructive);
}

.ued-icon-button--error:hover:not(:disabled) {
  border-color: #ffb8b0;
  background: var(--ued-color-error-soft);
}

@keyframes ued-icon-spin {
  to {
    transform: rotate(360deg);
  }
}
</style>
