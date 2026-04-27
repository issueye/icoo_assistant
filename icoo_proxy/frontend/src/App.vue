<template>
  <div class="app-shell">
    <header class="app-global-header">
      <div class="app-global-header__brand">
        <div class="app-global-header__logo" aria-hidden="true">
          <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polygon points="12 2 2 7 12 12 22 7 12 2"/><polyline points="2 17 12 22 22 17"/><polyline points="2 12 12 17 22 12"/></svg>
        </div>
        <div class="app-global-header__title">
          <span>本地 AI 网关管理台</span>
        </div>
      </div>
      <div class="app-global-header__right">
        <div class="app-window-controls" aria-label="窗口控制">
          <button class="app-window-control" type="button" aria-label="最小化" title="最小化" @click="minimizeWindow">
            <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M5 12h14"/></svg>
          </button>
          <button class="app-window-control app-window-control--close" type="button" aria-label="关闭" title="关闭" @click="closeWindow">
            <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M18 6 6 18"/><path d="m6 6 12 12"/></svg>
          </button>
        </div>
      </div>
    </header>

    <div class="app-frame">
      <aside class="app-sidebar">
        <nav class="app-sidebar-nav">
          <RouterLink
            v-for="item in navItems"
            :key="item.to"
            :to="item.to"
            class="app-nav-item"
            :class="route.path === item.to ? 'app-nav-item-active' : 'app-nav-item-idle'"
          >
            <span class="app-nav-icon" aria-hidden="true">
              <svg v-if="item.icon === 'overview'" xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="3" width="7" height="7"/><rect x="14" y="3" width="7" height="7"/><rect x="14" y="14" width="7" height="7"/><rect x="3" y="14" width="7" height="7"/></svg>
              <svg v-else-if="item.icon === 'supplier'" xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="2" y="2" width="20" height="8" rx="2" ry="2"/><rect x="2" y="14" width="20" height="8" rx="2" ry="2"/><line x1="6" y1="6" x2="6.01" y2="6"/><line x1="6" y1="18" x2="6.01" y2="18"/></svg>
              <svg v-else-if="item.icon === 'model'" xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 2H2v10l9.29 9.29c.94.94 2.48.94 3.42 0l6.58-6.58c.94-.94.94-2.48 0-3.42L12 2Z"/><path d="M7 7h.01"/></svg>
              <svg v-else-if="item.icon === 'endpoint'" xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><path d="M2 12h20"/><path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z"/></svg>
              <svg v-else-if="item.icon === 'key'" xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="m21 2-2 2m-7.61 7.61a5.5 5.5 0 1 1-7.778 7.778 5.5 5.5 0 0 1 7.777-7.777zm0 0L15.5 7.5m0 0 3 3L22 7l-3-3m-3.5 3.5L19 4"/></svg>
              <svg v-else-if="item.icon === 'traffic'" xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="22 12 18 12 15 21 9 3 6 12 2 12"/></svg>
              <svg v-else-if="item.icon === 'settings'" xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12.22 2h-.44a2 2 0 0 0-2 2v.18a2 2 0 0 1-1 1.73l-.43.25a2 2 0 0 1-2 0l-.15-.08a2 2 0 0 0-2.73.73l-.22.38a2 2 0 0 0 .73 2.73l.15.1a2 2 0 0 1 1 1.72v.51a2 2 0 0 1-1 1.74l-.15.09a2 2 0 0 0-.73 2.73l.22.38a2 2 0 0 0 2.73.73l.15-.08a2 2 0 0 1 2 0l.43.25a2 2 0 0 1 1 1.73V20a2 2 0 0 0 2 2h.44a2 2 0 0 0 2-2v-.18a2 2 0 0 1 1-1.73l.43-.25a2 2 0 0 1 2 0l.15.08a2 2 0 0 0 2.73-.73l.22-.39a2 2 0 0 0-.73-2.73l-.15-.08a2 2 0 0 1-1-1.74v-.5a2 2 0 0 1 1-1.74l.15-.09a2 2 0 0 0 .73-2.73l-.22-.38a2 2 0 0 0-2.73-.73l-.15.08a2 2 0 0 1-2 0l-.43-.25a2 2 0 0 1-1-1.73V4a2 2 0 0 0-2-2z"/><circle cx="12" cy="12" r="3"/></svg>
              <svg v-else xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8Z"/><path d="M14 2v6h6"/><path d="M16 13H8"/><path d="M16 17H8"/><path d="M10 9H8"/></svg>
            </span>
            <span class="app-nav-text">{{ item.label }}</span>
          </RouterLink>
        </nav>
      </aside>

      <main class="app-main">
        <header class="app-topbar">
          <div class="min-w-0">
            <p class="truncate text-sm font-semibold text-[#262626]">{{ currentTitle }}</p>
          </div>
          <div id="app-topbar-actions" class="app-topbar-actions" />
        </header>

        <div class="app-content">
          <RouterView />
        </div>
      </main>
    </div>

    <footer class="app-global-footer">
      <span>icoo proxy</span>
      <span>本地代理服务由桌面端自动托管</span>
    </footer>

    <UMessage />
  </div>
</template>

<script setup>
import { computed } from "vue";
import { RouterLink, RouterView, useRoute } from "vue-router";
import { Quit, WindowMinimise } from "../wailsjs/runtime/runtime";
import UMessage from "./components/ued/UMessage.vue";

const route = useRoute();

const navItems = computed(() => [
  { to: "/", label: "网关概览", icon: "overview" },
  { to: "/suppliers", label: "供应商", icon: "supplier" },
  { to: "/model-aliases", label: "模型别名", icon: "model" },
  { to: "/endpoints", label: "端点", icon: "endpoint" },
  { to: "/auth-keys", label: "授权 Key", icon: "key" },
  { to: "/traffic", label: "流量", icon: "traffic" },
  { to: "/settings", label: "设置", icon: "settings" },
  { to: "/ued", label: "UED", icon: "doc" },
]);

const currentTitle = computed(() => {
  const current = navItems.value.find((item) => item.to === route.path);
  return current?.label || "本地 AI 网关管理台";
});

function minimizeWindow() {
  WindowMinimise();
}

function closeWindow() {
  Quit();
}
</script>
