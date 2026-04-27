<template>
  <Teleport to="body">
    <div v-if="open" class="ued-modal" role="dialog" aria-modal="true" :aria-labelledby="titleId" @keydown.esc="close">
      <div class="ued-modal__mask" @click="handleMaskClick" />
      <div ref="panelRef" class="ued-modal__panel" :style="{ width }" tabindex="-1">
        <div class="ued-modal__header">
          <h3 :id="titleId" class="ued-modal__title">{{ title }}</h3>
          <button class="ued-modal__close" type="button" aria-label="关闭弹窗" @click="close">
            <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M18 6 6 18" /><path d="m6 6 12 12" /></svg>
          </button>
        </div>
        <div class="ued-modal__body">
          <slot />
        </div>
        <div v-if="$slots.footer" class="ued-modal__footer">
          <slot name="footer" />
        </div>
      </div>
    </div>
  </Teleport>
</template>

<script setup>
import { nextTick, ref, watch } from "vue";

const emit = defineEmits(["update:open", "close"]);

const props = defineProps({
  open: {
    type: Boolean,
    default: false,
  },
  title: {
    type: String,
    required: true,
  },
  width: {
    type: String,
    default: "520px",
  },
  closeOnMask: {
    type: Boolean,
    default: true,
  },
});

const titleId = `ued-modal-title-${Math.random().toString(36).slice(2, 9)}`;
const panelRef = ref(null);

watch(
  () => props.open,
  async (value) => {
    if (!value) {
      return;
    }
    await nextTick();
    panelRef.value?.focus();
  },
);

function close() {
  emit("update:open", false);
  emit("close");
}

function handleMaskClick() {
  if (props.closeOnMask) {
    close();
  }
}
</script>
