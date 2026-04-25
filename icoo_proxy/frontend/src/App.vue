<template>
  <div class="app-shell">
    <div class="app-frame">
      <aside class="app-sidebar">
        <div class="app-sidebar-brand">
          <h1 class="text-lg font-semibold text-zinc-900">icoo proxy</h1>
          <p class="mt-1 text-xs text-zinc-500">本地 AI 网关</p>
        </div>

        <nav class="app-sidebar-nav">
          <RouterLink
            v-for="item in navItems"
            :key="item.to"
            :to="item.to"
            class="app-nav-item"
            :class="route.path === item.to ? 'app-nav-item-active' : 'app-nav-item-idle'"
          >
            <p class="font-medium">{{ item.label }}</p>
          </RouterLink>
        </nav>

        <div class="app-sidebar-footer">
          <p class="text-xs font-medium text-zinc-500">轻量管理台</p>
        </div>
      </aside>

      <main class="app-main">
        <header class="app-topbar">
          <div>
            <p class="text-base font-semibold text-zinc-900">{{ currentTitle }}</p>
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
    label: "网关概览",
  },
  {
    to: "/suppliers",
    label: "供应商管理",
  },
  {
    to: "/endpoints",
    label: "端点管理",
  },
  {
    to: "/auth-keys",
    label: "授权 Key",
  },
  {
    to: "/traffic",
    label: "流量监控",
  },
  {
    to: "/ued",
    label: "UED 规范",
  },
]);

const currentTitle = computed(() => {
  const current = navItems.value.find((item) => item.to === route.path);
  return current?.label || "本地 AI 网关管理台";
});
</script>
