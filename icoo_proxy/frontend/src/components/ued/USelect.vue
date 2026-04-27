<template>
  <UFormField :label="label" :hint="hint" :error="error" :required="required">
    <div ref="rootRef" class="ued-select" :class="{ 'is-open': open, 'is-disabled': disabled, 'is-error': error }">
      <button
        type="button"
        class="ued-select__control"
        :disabled="disabled"
        :aria-expanded="open"
        :aria-controls="listboxId"
        aria-haspopup="listbox"
        @click="toggle"
        @keydown="handleControlKeydown"
      >
        <span class="ued-select__value" :class="{ 'is-placeholder': !selectedOption }">
          {{ selectedOption?.label || placeholder }}
        </span>
        <span class="ued-select__arrow" aria-hidden="true">
          <svg xmlns="http://www.w3.org/2000/svg" width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round"><path d="m6 9 6 6 6-6" /></svg>
        </span>
      </button>

      <div v-if="open" :id="listboxId" class="ued-select__menu" role="listbox" tabindex="-1">
        <button
          v-for="(option, index) in normalizedOptions"
          :key="option.value"
          type="button"
          class="ued-select__option"
          :class="{ 'is-selected': option.value === modelValue, 'is-active': index === activeIndex }"
          role="option"
          :aria-selected="option.value === modelValue"
          @mouseenter="activeIndex = index"
          @click="choose(option)"
        >
          <span>{{ option.label }}</span>
          <span v-if="option.value === modelValue" class="ued-select__check" aria-hidden="true">
            <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.4" stroke-linecap="round" stroke-linejoin="round"><path d="M20 6 9 17l-5-5" /></svg>
          </span>
        </button>
        <div v-if="normalizedOptions.length === 0" class="ued-select__empty">
          暂无选项
        </div>
      </div>
    </div>
  </UFormField>
</template>

<script setup>
import { computed, onBeforeUnmount, onMounted, ref, watch } from "vue";
import UFormField from "./UFormField.vue";

const emit = defineEmits(["update:modelValue", "change"]);

const props = defineProps({
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
    default: "请选择",
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
  options: {
    type: Array,
    default: () => [],
  },
});

const rootRef = ref(null);
const open = ref(false);
const activeIndex = ref(-1);
const listboxId = `ued-select-${Math.random().toString(36).slice(2, 9)}`;

const normalizedOptions = computed(() =>
  props.options.map((option) => {
    if (typeof option === "string" || typeof option === "number") {
      return { label: String(option), value: option };
    }
    return {
      label: option.label ?? String(option.value ?? ""),
      value: option.value ?? "",
    };
  }),
);

const selectedOption = computed(() => normalizedOptions.value.find((option) => option.value === props.modelValue));

watch(open, (value) => {
  if (!value) {
    return;
  }
  const selectedIndex = normalizedOptions.value.findIndex((option) => option.value === props.modelValue);
  activeIndex.value = selectedIndex >= 0 ? selectedIndex : 0;
});

function toggle() {
  if (props.disabled) {
    return;
  }
  open.value = !open.value;
}

function close() {
  open.value = false;
}

function choose(option) {
  emit("update:modelValue", option.value);
  emit("change", option.value);
  close();
}

function moveActive(step) {
  if (!open.value) {
    open.value = true;
    return;
  }
  const count = normalizedOptions.value.length;
  if (count === 0) {
    activeIndex.value = -1;
    return;
  }
  activeIndex.value = (activeIndex.value + step + count) % count;
}

function handleControlKeydown(event) {
  switch (event.key) {
    case "ArrowDown":
      event.preventDefault();
      moveActive(1);
      break;
    case "ArrowUp":
      event.preventDefault();
      moveActive(-1);
      break;
    case "Enter":
    case " ":
      event.preventDefault();
      if (open.value && activeIndex.value >= 0) {
        choose(normalizedOptions.value[activeIndex.value]);
      } else {
        open.value = true;
      }
      break;
    case "Escape":
      close();
      break;
    default:
      break;
  }
}

function handleDocumentClick(event) {
  if (!rootRef.value?.contains(event.target)) {
    close();
  }
}

onMounted(() => {
  document.addEventListener("click", handleDocumentClick);
});

onBeforeUnmount(() => {
  document.removeEventListener("click", handleDocumentClick);
});
</script>
