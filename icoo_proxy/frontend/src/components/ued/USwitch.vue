<template>
  <label class="ued-switch" :class="{ 'is-disabled': disabled }">
    <input
      class="sr-only"
      type="checkbox"
      :checked="modelValue"
      :disabled="disabled"
      @change="$emit('update:modelValue', $event.target.checked)"
    />
    <span class="ued-switch__track" :class="{ 'is-on': modelValue }">
      <span class="ued-switch__thumb" />
    </span>
    <span class="ued-switch__content">
      <span class="ued-switch__label">{{ label }}</span>
      <span v-if="hint" class="ued-switch__hint">{{ hint }}</span>
    </span>
  </label>
</template>

<script setup>
defineEmits(["update:modelValue"]);

defineProps({
  modelValue: {
    type: Boolean,
    default: false,
  },
  label: {
    type: String,
    required: true,
  },
  hint: {
    type: String,
    default: "",
  },
  disabled: {
    type: Boolean,
    default: false,
  },
});
</script>
