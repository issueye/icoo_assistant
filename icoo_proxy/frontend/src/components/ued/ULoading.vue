<template>
  <Teleport v-if="fullscreen" to="body">
    <Transition name="ued-loading-fade">
      <div v-if="visible" class="ued-loading-fullscreen" role="status" aria-live="polite">
        <LoadingIndicator />
      </div>
    </Transition>
  </Teleport>

  <div v-else-if="$slots.default" class="ued-loading-container" :class="{ 'is-loading': visible }">
    <slot />
    <Transition name="ued-loading-fade">
      <div v-if="visible" class="ued-loading-overlay" role="status" aria-live="polite">
        <LoadingIndicator />
      </div>
    </Transition>
  </div>

  <LoadingIndicator v-else-if="visible" />
</template>

<script setup>
import { computed, h, ref, watch } from "vue";

const props = defineProps({
  spinning: {
    type: Boolean,
    default: true,
  },
  tip: {
    type: String,
    default: "",
  },
  size: {
    type: String,
    default: "md",
  },
  fullscreen: {
    type: Boolean,
    default: false,
  },
  delay: {
    type: Number,
    default: 0,
  },
});

const visible = ref(false);
let timer = null;

const normalizedSize = computed(() => {
  const value = String(props.size || "md").toLowerCase();
  return ["sm", "md", "lg"].includes(value) ? value : "md";
});

const LoadingIndicator = () =>
  h(
    "div",
    {
      class: ["ued-loading", `ued-loading--${normalizedSize.value}`],
    },
    [
      h("span", { class: "ued-loading__spinner", "aria-hidden": "true" }),
      props.tip ? h("span", { class: "ued-loading__tip" }, props.tip) : null,
    ],
  );

watch(
  () => props.spinning,
  (spinning) => {
    clearTimeout(timer);
    if (!spinning) {
      visible.value = false;
      return;
    }

    if (props.delay > 0) {
      timer = window.setTimeout(() => {
        visible.value = true;
      }, props.delay);
      return;
    }

    visible.value = true;
  },
  { immediate: true },
);
</script>
