<template>
  <div
    ref="triggerRef"
    class="ued-tooltip"
    @mouseenter="handleMouseEnter"
    @mouseleave="handleMouseLeave"
  >
    <slot />
    <Transition name="ued-tooltip">
      <div
        v-if="visible"
        ref="tooltipRef"
        class="ued-tooltip__popup"
        :class="[`ued-tooltip__popup--${actualPlacement}`]"
        :style="popupStyle"
      >
        <div class="ued-tooltip__arrow"></div>
        <div class="ued-tooltip__inner">
          <slot name="content">{{ content }}</slot>
        </div>
      </div>
    </Transition>
  </div>
</template>

<script setup>
import { computed, nextTick, ref, watch } from "vue";

const props = defineProps({
  content: { type: String, default: "" },
  placement: { type: String, default: "top" },
  disabled: { type: Boolean, default: false },
  delay: { type: Number, default: 200 },
});

const triggerRef = ref(null);
const tooltipRef = ref(null);
const visible = ref(false);
const actualPlacement = ref(props.placement);
const popupOffset = ref({ top: 0, left: 0 });
let enterTimer = null;

const popupStyle = computed(() => ({
  top: popupOffset.value.top + "px",
  left: popupOffset.value.left + "px",
}));

function handleMouseEnter() {
  if (props.disabled) return;
  clearTimeout(enterTimer);
  enterTimer = setTimeout(() => {
    visible.value = true;
  }, props.delay);
}

function handleMouseLeave() {
  clearTimeout(enterTimer);
  visible.value = false;
}

watch(visible, async (val) => {
  if (!val) return;
  await nextTick();
  updatePosition();
});

function updatePosition() {
  const trigger = triggerRef.value;
  const popup = tooltipRef.value;
  if (!trigger || !popup) return;

  const triggerRect = trigger.getBoundingClientRect();
  const popupRect = popup.getBoundingClientRect();
  const spacing = 8;

  let top = 0;
  let left = 0;
  let placement = props.placement;

  // Try requested placement first
  const fitsAbove = triggerRect.top >= popupRect.height + spacing;
  const fitsBelow =
    window.innerHeight - triggerRect.bottom >= popupRect.height + spacing;
  const fitsLeft = triggerRect.left >= popupRect.width + spacing;
  const fitsRight =
    window.innerWidth - triggerRect.right >= popupRect.width + spacing;

  if (placement === "top" && !fitsAbove && fitsBelow) placement = "bottom";
  if (placement === "bottom" && !fitsBelow && fitsAbove) placement = "top";

  actualPlacement.value = placement;

  switch (placement) {
    case "top":
      top = -popupRect.height - spacing;
      left = (triggerRect.width - popupRect.width) / 2;
      break;
    case "bottom":
      top = triggerRect.height + spacing;
      left = (triggerRect.width - popupRect.width) / 2;
      break;
    case "left":
      top = (triggerRect.height - popupRect.height) / 2;
      left = -popupRect.width - spacing;
      break;
    case "right":
      top = (triggerRect.height - popupRect.height) / 2;
      left = triggerRect.width + spacing;
      break;
  }

  popupOffset.value = { top, left };
}
</script>

<style scoped>
.ued-tooltip {
  position: relative;
  display: inline-flex;
  max-width: 100%;
}

.ued-tooltip__popup {
  position: absolute;
  z-index: 100;
  pointer-events: none;
}

.ued-tooltip__inner {
  max-width: 320px;
  padding: 7px 10px;
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 6px;
  background: #172033;
  color: #f8fafc;
  font-size: 12px;
  line-height: 1.5;
  word-break: break-word;
  box-shadow: 0 14px 32px rgba(20, 31, 50, 0.22), 0 4px 10px rgba(20, 31, 50, 0.12);
}

.ued-tooltip__arrow {
  position: absolute;
  width: 6px;
  height: 6px;
  background: #172033;
  transform: rotate(45deg);
}

.ued-tooltip__popup--top .ued-tooltip__arrow {
  bottom: -3px;
  left: 50%;
  margin-left: -3px;
}

.ued-tooltip__popup--bottom .ued-tooltip__arrow {
  top: -3px;
  left: 50%;
  margin-left: -3px;
}

.ued-tooltip__popup--left .ued-tooltip__arrow {
  right: -3px;
  top: 50%;
  margin-top: -3px;
}

.ued-tooltip__popup--right .ued-tooltip__arrow {
  left: -3px;
  top: 50%;
  margin-top: -3px;
}

.ued-tooltip-enter-active,
.ued-tooltip-leave-active {
  transition: opacity 0.2s ease, transform 0.2s ease;
}

.ued-tooltip-enter-from,
.ued-tooltip-leave-to {
  opacity: 0;
  transform: scale(0.96);
}
</style>
