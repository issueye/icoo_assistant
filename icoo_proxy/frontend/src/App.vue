<template>
  <div class="app-shell" :class="{ 'app-shell--sidebar-collapsed': sidebarCollapsed }">
    <header class="app-titlebar">
      <div class="app-titlebar__brand">
        <img class="app-titlebar__logo" src="./assets/images/appicon.png" alt="" />
        <span class="app-titlebar__name">icoo_proxy</span>
        <span class="app-titlebar__divider" aria-hidden="true"></span>
        <span class="app-titlebar__caption">本地 AI 协议转换网关</span>
      </div>

      <div class="app-window-controls" aria-label="窗口控制">
        <button class="app-window-control" type="button" aria-label="最小化" title="最小化" @click="minimizeWindow">
          <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M5 12h14" /></svg>
        </button>
        <button class="app-window-control app-window-control--close" type="button" aria-label="关闭" title="关闭" @click="closeWindow">
          <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M18 6 6 18" /><path d="m6 6 12 12" /></svg>
        </button>
      </div>
    </header>

    <div class="app-workbench">
      <aside class="app-sidebar">
        <div class="app-sidebar__header">
          <button
            class="app-sidebar-toggle"
            type="button"
            :aria-label="sidebarCollapsed ? '展开导航' : '收起导航'"
            :title="sidebarCollapsed ? '展开导航' : '收起导航'"
            @click="toggleSidebar"
          >
            <svg v-if="sidebarCollapsed" xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="4" x2="20" y1="12" y2="12" /><line x1="4" x2="20" y1="6" y2="6" /><line x1="4" x2="20" y1="18" y2="18" /></svg>
            <svg v-else xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M15 18l-6-6 6-6" /></svg>
          </button>
          <div class="app-sidebar__heading">
            <span>控制台</span>
          </div>
        </div>

        <nav class="app-sidebar-nav" aria-label="主导航">
          <section v-for="group in navGroups" :key="group.name" class="app-nav-group">
            <p class="app-nav-group__label">{{ group.name }}</p>
            <RouterLink
              v-for="item in group.items"
              :key="item.to"
              :to="item.to"
              class="app-nav-item"
              :class="{ 'app-nav-item--active': route.path === item.to }"
              :title="sidebarCollapsed ? item.label : item.description"
            >
              <span class="app-nav-icon" aria-hidden="true">
                <svg v-if="item.icon === 'overview'" xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="3" width="7" height="7" /><rect x="14" y="3" width="7" height="7" /><rect x="14" y="14" width="7" height="7" /><rect x="3" y="14" width="7" height="7" /></svg>
                <svg v-else-if="item.icon === 'supplier'" xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="2" y="2" width="20" height="8" rx="2" /><rect x="2" y="14" width="20" height="8" rx="2" /><line x1="6" y1="6" x2="6.01" y2="6" /><line x1="6" y1="18" x2="6.01" y2="18" /></svg>
                <svg v-else-if="item.icon === 'model'" xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12 2H2v10l9.29 9.29c.94.94 2.48.94 3.42 0l6.58-6.58c.94-.94.94-2.48 0-3.42L12 2Z" /><path d="M7 7h.01" /></svg>
                <svg v-else-if="item.icon === 'endpoint'" xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10" /><path d="M2 12h20" /><path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z" /></svg>
                <svg v-else-if="item.icon === 'key'" xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="m21 2-2 2m-7.61 7.61a5.5 5.5 0 1 1-7.778 7.778 5.5 5.5 0 0 1 7.777-7.777zm0 0L15.5 7.5m0 0 3 3L22 7l-3-3m-3.5 3.5L19 4" /></svg>
                <svg v-else-if="item.icon === 'traffic'" xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="22 12 18 12 15 21 9 3 6 12 2 12" /></svg>
                <svg v-else-if="item.icon === 'settings'" xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M12.22 2h-.44a2 2 0 0 0-2 2v.18a2 2 0 0 1-1 1.73l-.43.25a2 2 0 0 1-2 0l-.15-.08a2 2 0 0 0-2.73.73l-.22.38a2 2 0 0 0 .73 2.73l.15.1a2 2 0 0 1 1 1.72v.51a2 2 0 0 1-1 1.74l-.15.09a2 2 0 0 0-.73 2.73l.22.38a2 2 0 0 0 2.73.73l.15-.08a2 2 0 0 1 2 0l.43.25a2 2 0 0 1 1 1.73V20a2 2 0 0 0 2 2h.44a2 2 0 0 0 2-2v-.18a2 2 0 0 1 1-1.73l.43-.25a2 2 0 0 1 2 0l.15.08a2 2 0 0 0 2.73-.73l.22-.39a2 2 0 0 0-.73-2.73l-.15-.08a2 2 0 0 1-1-1.74v-.5a2 2 0 0 1 1-1.74l.15-.09a2 2 0 0 0 .73-2.73l-.22-.38a2 2 0 0 0-2.73-.73l-.15.08a2 2 0 0 1-2 0l-.43-.25a2 2 0 0 1-1-1.73V4a2 2 0 0 0-2-2z" /><circle cx="12" cy="12" r="3" /></svg>
                <svg v-else-if="item.icon === 'ued'" xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M4 4h16v16H4z" /><path d="M4 9h16" /><path d="M9 20V9" /></svg>
              </span>
              <span class="app-nav-copy">
                <span class="app-nav-text">{{ item.label }}</span>
                <span class="app-nav-desc">{{ item.description }}</span>
              </span>
            </RouterLink>
          </section>
        </nav>

        <div class="app-sidebar__status">
          <span class="app-status-dot" :class="proxyStatusDotClass" aria-hidden="true"></span>
        </div>
      </aside>

      <main class="app-main">
        <header class="app-topbar">
          <div class="app-page-identity">
            <div class="app-breadcrumb">
              <span>icoo_proxy</span>
              <span>/</span>
              <span>{{ currentTitle }}</span>
            </div>
          </div>
          <div id="app-topbar-actions" class="app-topbar-actions" />
        </header>

        <div class="app-content">
          <RouterView />
        </div>
      </main>
    </div>

    <footer class="app-statusbar">
      <span>Ready</span>
      <span class="app-statusbar__item">Wails Desktop</span>
      <span class="app-statusbar__item">{{ proxyStatusDetail }}</span>
      <span class="app-statusbar__item app-statusbar__item--right">代理{{ proxyStatusText }}</span>
    </footer>

    <UMessage />
  </div>
</template>

<script setup>
import { computed, onMounted, onUnmounted, ref } from "vue";
import { RouterLink, RouterView, useRoute } from "vue-router";
import { Quit, WindowMinimise } from "../wailsjs/runtime/runtime";
import { State } from "./lib/wailsApp";
import UMessage from "./components/ued/UMessage.vue";

const route = useRoute();
const sidebarCollapsed = ref(false);
const proxyState = ref(null);
let proxyStateTimer = null;

const navGroups = computed(() => [
  {
    name: "网关",
    items: [
      { to: "/", label: "网关概览", description: "", icon: "overview" },
      { to: "/traffic", label: "流量监控", description: "", icon: "traffic" },
    ],
  },
  {
    name: "配置",
    items: [
      { to: "/suppliers", label: "供应商", description: "", icon: "supplier" },
      { to: "/model-aliases", label: "模型别名", description: "", icon: "model" },
      { to: "/endpoints", label: "端点", description: "", icon: "endpoint" },
      { to: "/auth-keys", label: "授权 Key", description: "", icon: "key" },
      { to: "/settings", label: "项目设置", description: "", icon: "settings" },
    ],
  },
  {
    name: "UED",
    items: [
      { to: "/ued", label: "组件规范", description: "", icon: "ued" },
    ],
  },
]);

const navItems = computed(() => navGroups.value.flatMap((group) => group.items));

const currentNavItem = computed(() => navItems.value.find((item) => item.to === route.path));

const currentTitle = computed(() => currentNavItem.value?.label || "本地 AI 网关管理台");

const proxyStatusText = computed(() => {
  if (!proxyState.value) {
    return "检测中";
  }
  if (proxyState.value.last_error) {
    return "异常";
  }
  return proxyState.value.running ? "运行中" : "已停止";
});

const proxyStatusDetail = computed(() => {
  if (!proxyState.value) {
    return "正在读取代理状态";
  }
  if (proxyState.value.last_error) {
    return proxyState.value.last_error;
  }
  return proxyState.value.proxy_url || proxyState.value.listen_addr || "未监听";
});

const proxyStatusDotClass = computed(() => ({
  "app-status-dot--running": proxyState.value?.running && !proxyState.value?.last_error,
  "app-status-dot--stopped": proxyState.value && !proxyState.value.running && !proxyState.value.last_error,
  "app-status-dot--error": Boolean(proxyState.value?.last_error),
}));

function toggleSidebar() {
  sidebarCollapsed.value = !sidebarCollapsed.value;
}

async function refreshProxyState() {
  try {
    proxyState.value = await State();
  } catch (error) {
    proxyState.value = {
      running: false,
      last_error: error instanceof Error ? error.message : String(error),
    };
  }
}

onMounted(() => {
  refreshProxyState();
  proxyStateTimer = window.setInterval(refreshProxyState, 5000);
});

onUnmounted(() => {
  if (proxyStateTimer) {
    window.clearInterval(proxyStateTimer);
  }
});

function minimizeWindow() {
  WindowMinimise();
}

function closeWindow() {
  Quit();
}
</script>
