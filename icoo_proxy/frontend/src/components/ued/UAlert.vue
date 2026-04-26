<template>
  <div
    v-if="visible"
    class="ued-alert"
    :class="[
      `ued-alert--${normalizedType}`,
      {
        'ued-alert--with-description': Boolean(description || $slots.description),
        'ued-alert--banner': banner,
      },
    ]"
    role="alert"
  >
    <span v-if="showIcon" class="ued-alert__icon" aria-hidden="true">
      {{ iconText }}
    </span>
    <div class="ued-alert__content">
      <div class="ued-alert__message">
        <slot>{{ message }}</slot>
      </div>
      <div v-if="description || $slots.description" class="ued-alert__description">
        <slot name="description">{{ description }}</slot>
      </div>
    </div>
    <div v-if="$slots.action" class="ued-alert__action">
      <slot name="action" />
    </div>
    <button
      v-if="closable"
      type="button"
      class="ued-alert__close"
      aria-label="关闭提示"
      @click="close"
    >
      ×
    </button>
  </div>
</template>

<script setup>
import { computed, ref, watch } from "vue";

const emit = defineEmits(["close", "update:open"]);

const props = defineProps({
  type: {
    type: String,
    default: "info",
  },
  message: {
    type: String,
    default: "",
  },
  description: {
    type: String,
    default: "",
  },
  showIcon: {
    type: Boolean,
    default: true,
  },
  closable: {
    type: Boolean,
    default: false,
  },
  banner: {
    type: Boolean,
    default: false,
  },
  open: {
    type: Boolean,
    default: true,
  },
});

const visible = ref(props.open);

const normalizedType = computed(() => {
  const value = String(props.type || "info").toLowerCase();
  if (value === "danger") {
    return "error";
  }
  if (value === "sucess") {
    return "success";
  }
  return ["success", "info", "warning", "error"].includes(value) ? value : "info";
});

const iconText = computed(() => {
  const icons = {
    success: "✓",
    info: "i",
    warning: "!",
    error: "×",
  };
  return icons[normalizedType.value];
});

watch(
  () => props.open,
  (value) => {
    visible.value = value;
  },
);

function close() {
  visible.value = false;
  emit("update:open", false);
  emit("close");
}
</script>
