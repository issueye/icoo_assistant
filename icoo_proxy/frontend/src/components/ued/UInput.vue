<template>
  <UFormField :label="label" :hint="hint" :error="error" :required="required">
    <textarea
      v-if="textarea"
      class="ued-input ued-input--textarea"
      :class="{ 'is-error': error }"
      :value="modelValue"
      :placeholder="placeholder"
      :disabled="disabled"
      :rows="rows"
      @input="$emit('update:modelValue', $event.target.value)"
    />
    <input
      v-else
      class="ued-input"
      :class="{ 'is-error': error }"
      :type="type"
      :value="modelValue"
      :placeholder="placeholder"
      :disabled="disabled"
      @input="$emit('update:modelValue', $event.target.value)"
    />
  </UFormField>
</template>

<script setup>
import UFormField from "./UFormField.vue";

defineEmits(["update:modelValue"]);

defineProps({
  modelValue: {
    type: [String, Number],
    default: "",
  },
  label: {
    type: String,
    required: true,
  },
  placeholder: {
    type: String,
    default: "",
  },
  hint: {
    type: String,
    default: "",
  },
  error: {
    type: String,
    default: "",
  },
  required: {
    type: Boolean,
    default: false,
  },
  disabled: {
    type: Boolean,
    default: false,
  },
  textarea: {
    type: Boolean,
    default: false,
  },
  rows: {
    type: Number,
    default: 4,
  },
  type: {
    type: String,
    default: "text",
  },
});
</script>
