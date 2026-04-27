<template>
  <article class="stat-card" :class="`stat-card--${tone}`">
    <div class="stat-card__icon" aria-hidden="true">
      <slot name="icon">
        <svg v-if="icon === 'globe'" xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><path d="M2 12h20"/><path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z"/></svg>
        <svg v-else-if="icon === 'wifi'" xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M5 12.55a11 11 0 0 1 14.08 0"/><path d="M1.42 9a16 16 0 0 1 21.16 0"/><path d="M8.53 16.11a6 6 0 0 1 6.95 0"/><line x1="12" y1="20" x2="12.01" y2="20"/></svg>
        <svg v-else-if="icon === 'key'" xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="m21 2-2 2m-7.61 7.61a5.5 5.5 0 1 1-7.778 7.778 5.5 5.5 0 0 1 7.777-7.777zm0 0L15.5 7.5m0 0 3 3L22 7l-3-3m-3.5 3.5L19 4"/></svg>
        <svg v-else-if="icon === 'tag'" xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 2H2v10l9.29 9.29c.94.94 2.48.94 3.42 0l6.58-6.58c.94-.94.94-2.48 0-3.42L12 2Z"/><path d="M7 7h.01"/></svg>
        <svg v-else-if="icon === 'server'" xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="2" y="2" width="20" height="8" rx="2" ry="2"/><rect x="2" y="14" width="20" height="8" rx="2" ry="2"/><line x1="6" y1="6" x2="6.01" y2="6"/><line x1="6" y1="18" x2="6.01" y2="18"/></svg>
        <svg v-else-if="icon === 'check'" xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/><polyline points="22 4 12 14.01 9 11.01"/></svg>
        <svg v-else-if="icon === 'activity'" xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="22 12 18 12 15 21 9 3 6 12 2 12"/></svg>
        <svg v-else-if="icon === 'layers'" xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polygon points="12 2 2 7 12 12 22 7 12 2"/><polyline points="2 17 12 22 22 17"/><polyline points="2 12 12 17 22 12"/></svg>
        <svg v-else-if="icon === 'heart-pulse'" xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M19 14c1.49-1.46 3-3.21 3-5.5A5.5 5.5 0 0 0 16.5 3c-1.76 0-3 .5-4.5 2-1.5-1.5-2.74-2-4.5-2A5.5 5.5 0 0 0 2 8.5c0 2.3 1.5 4.05 3 5.5l7 7Z"/><path d="M3.22 12H9.5l.5-1 2 4.5 2-7 1.5 3.5h5.27"/></svg>
        <svg v-else xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="3" width="18" height="18" rx="2" ry="2"/></svg>
      </slot>
    </div>
    <div class="stat-card__body">
      <p class="stat-card__label">{{ label }}</p>
      <p class="stat-card__value">{{ value }}</p>
    </div>
  </article>
</template>

<script setup>
defineProps({
  label: {
    type: String,
    required: true,
  },
  value: {
    type: String,
    required: true,
  },
  icon: {
    type: String,
    default: "",
  },
  tone: {
    type: String,
    default: "neutral",
    validator: (v) => ["neutral", "primary", "success", "warning", "danger", "info"].includes(v),
  },
});
</script>

<style scoped>
.stat-card {
  @apply relative flex min-h-[84px] items-center gap-3 rounded border bg-white px-3 py-3 transition;
  border-color: var(--ued-color-border);
  border-radius: var(--ued-radius-md);
  box-shadow: var(--ued-shadow-card);
}
.stat-card:hover {
  border-color: #b8c7f6;
  box-shadow: var(--ued-shadow-card-hover);
  transform: translateY(-1px);
}
.stat-card__icon {
  @apply inline-flex h-10 w-10 shrink-0 items-center justify-center rounded border bg-slate-50 text-slate-500;
  border-color: var(--ued-color-border);
  border-radius: var(--ued-radius-md);
}
.stat-card__body {
  @apply min-w-0 flex-1;
}
.stat-card__label {
  @apply text-xs font-medium;
  color: var(--ued-color-text-muted);
}
.stat-card__value {
  @apply mt-1 break-all text-base font-semibold leading-5;
  color: var(--ued-color-text);
}

/* Tones */
.stat-card--primary .stat-card__icon {
  border-color: #c8d4ff;
  background: var(--ued-color-primary-soft);
  color: var(--ued-color-primary);
}
.stat-card--success .stat-card__icon {
  border-color: #bde8d3;
  background: var(--ued-color-success-soft);
  color: var(--ued-color-success);
}
.stat-card--warning .stat-card__icon {
  border-color: #f3d498;
  background: var(--ued-color-warning-soft);
  color: var(--ued-color-warning);
}
.stat-card--danger .stat-card__icon {
  border-color: #ffc5bf;
  background: var(--ued-color-error-soft);
  color: var(--ued-color-destructive);
}
.stat-card--info .stat-card__icon {
  border-color: #a6e5ef;
  background: var(--ued-color-info-soft);
  color: var(--ued-color-info);
}
</style>
