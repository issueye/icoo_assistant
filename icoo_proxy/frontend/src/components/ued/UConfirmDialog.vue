<template>
  <UModal :open="open" :title="title" width="420px" @update:open="handleOpen">
    <div class="space-y-2">
      <p class="text-sm leading-6 text-slate-600">{{ message }}</p>
      <p v-if="description" class="text-xs leading-6 text-slate-500">{{ description }}</p>
    </div>
    <template #footer>
      <div class="flex justify-end gap-2">
        <UButton variant="secondary" @click="handleCancel">{{ cancelText }}</UButton>
        <UButton :variant="danger ? 'error' : 'primary'" :loading="loading" @click="$emit('confirm')">
          {{ confirmText }}
        </UButton>
      </div>
    </template>
  </UModal>
</template>

<script setup>
import UButton from "./UButton.vue";
import UModal from "./UModal.vue";

const emit = defineEmits(["update:open", "confirm", "cancel"]);

defineProps({
  open: {
    type: Boolean,
    default: false,
  },
  title: {
    type: String,
    default: "确认操作",
  },
  message: {
    type: String,
    default: "",
  },
  description: {
    type: String,
    default: "",
  },
  confirmText: {
    type: String,
    default: "确认",
  },
  cancelText: {
    type: String,
    default: "取消",
  },
  loading: {
    type: Boolean,
    default: false,
  },
  danger: {
    type: Boolean,
    default: false,
  },
});

function handleOpen(value) {
  emit("update:open", value);
  if (!value) {
    emit("cancel");
  }
}

function handleCancel() {
  emit("update:open", false);
  emit("cancel");
}
</script>
