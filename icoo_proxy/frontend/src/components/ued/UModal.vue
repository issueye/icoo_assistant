<template>
  <Teleport to="body">
    <div v-if="open" class="ued-modal">
      <div class="ued-modal__mask" @click="handleMaskClick" />
      <div class="ued-modal__panel" :style="{ width }">
        <div class="ued-modal__header">
          <h3 class="ued-modal__title">{{ title }}</h3>
          <button class="ued-modal__close" type="button" @click="close">×</button>
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
