<template>
  <div class="app-shell">
    <div class="app-frame">
      <aside class="app-sidebar">
        <div class="app-sidebar-brand">
          <p class="text-xs font-semibold uppercase tracking-[0.16em] text-blue-700">icoo_proxy</p>
          <h1 class="mt-2 text-2xl font-semibold text-slate-900">AI Gateway Admin</h1>
          <p class="mt-2 text-sm leading-6 text-slate-500">
            Local desktop console for supplier management, protocol routing, and traffic inspection.
          </p>
        </div>

        <nav class="app-sidebar-nav">
          <RouterLink
            v-for="item in navItems"
            :key="item.to"
            :to="item.to"
            class="app-nav-item"
            :class="route.path === item.to ? 'app-nav-item-active' : 'app-nav-item-idle'"
          >
            <p class="text-sm font-medium">{{ item.label }}</p>
            <p class="mt-1 text-xs text-slate-500">{{ item.description }}</p>
          </RouterLink>
        </nav>

        <div class="app-sidebar-footer">
          <p class="text-xs font-semibold uppercase tracking-[0.14em] text-slate-500">Current Scope</p>
          <ul class="mt-3 space-y-2 text-sm text-slate-600">
            <li>Anthropic / OpenAI protocol entrypoints</li>
            <li>Non-streaming protocol translation</li>
            <li>Supplier, route policy, and traffic monitoring</li>
          </ul>
        </div>
      </aside>

      <main class="app-main">
        <header class="app-topbar">
          <div>
            <p class="text-xs font-semibold uppercase tracking-[0.14em] text-slate-500">Desktop Control Plane</p>
            <p class="mt-1 text-lg font-semibold text-slate-900">{{ currentTitle }}</p>
          </div>
          <div class="hidden items-center gap-2 md:flex">
            <span class="badge badge-info">Wails v2</span>
            <span class="badge badge-neutral">Vue 3 + Pinia</span>
          </div>
        </header>

        <div class="app-content">
          <RouterView />
        </div>
      </main>
    </div>
  </div>
</template>

<script setup>
import { computed } from "vue";
import { RouterLink, RouterView, useRoute } from "vue-router";

const route = useRoute();

const navItems = computed(() => [
  {
    to: "/",
    label: "Gateway Overview",
    description: "Runtime status, supplier health, and route summaries.",
  },
  {
    to: "/suppliers",
    label: "Supplier Management",
    description: "Manage supplier endpoints, health checks, and route policies.",
  },
  {
    to: "/traffic",
    label: "Traffic Monitor",
    description: "Review recent requests, failures, and response latency.",
  },
]);

const currentTitle = computed(() => {
  const current = navItems.value.find((item) => item.to === route.path);
  return current?.label || "AI Gateway Admin";
});
</script>
