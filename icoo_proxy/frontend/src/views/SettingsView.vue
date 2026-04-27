<template>
  <section class="page-section">
    <Teleport to="#app-topbar-actions">
      <div class="app-topbar-actions__group">
        <button
          class="btn btn-secondary"
          :class="{ 'is-loading': store.loading }"
          :disabled="store.loading || store.saving"
          @click="store.load"
        >
          <span v-if="store.loading" class="btn__spinner" />
          {{ store.loading ? "刷新中..." : "重新读取" }}
        </button>
        <button
          class="btn btn-primary"
          :class="{ 'is-loading': store.saving }"
          :disabled="store.loading || store.saving"
          @click="submit"
        >
          <span v-if="store.saving" class="btn__spinner" />
          {{ store.saving ? "保存中..." : "保存并重载" }}
        </button>
      </div>
    </Teleport>

    <div v-if="store.error" class="notice-error">
      {{ store.error }}
    </div>
    <div v-if="store.success" class="rounded-md border border-emerald-200 bg-emerald-50 px-3 py-2 text-sm text-emerald-700">
      {{ store.success }}
    </div>

    <PanelBlock title="外观设置">
      <div class="divide-y divide-[#f0f0f0]">
        <div class="py-3">
          <p class="text-sm font-medium text-[#262626]">主题颜色</p>
          <p class="mt-0.5 text-[11px] text-[#8c8c8c]">选择界面主色调，即时生效。</p>
          <div class="mt-3 flex flex-wrap gap-2">
            <button
              v-for="opt in uiPrefs.themeOptions"
              :key="opt.value"
              class="theme-swatch"
              :class="{ 'theme-swatch--active': uiPrefs.theme === opt.value }"
              :title="opt.label"
              @click="uiPrefs.setTheme(opt.value)"
            >
              <span class="theme-swatch__dot" :style="{ background: opt.color }" />
              <span class="theme-swatch__label">{{ opt.label }}</span>
            </button>
          </div>
        </div>

        <div class="py-3">
          <p class="text-sm font-medium text-[#262626]">按钮尺寸</p>
          <p class="mt-0.5 text-[11px] text-[#8c8c8c]">调整全局按钮和输入控件的大小。</p>
          <div class="mt-3 flex flex-wrap gap-2">
            <button
              v-for="opt in uiPrefs.buttonSizeOptions"
              :key="opt.value"
              class="btn"
              :class="uiPrefs.buttonSize === opt.value ? 'btn-primary' : 'btn-secondary'"
              @click="uiPrefs.setButtonSize(opt.value)"
            >
              {{ opt.label }}
            </button>
          </div>
          <div class="mt-3">
            <UButton size="sm">示例按钮 SM</UButton>
            <span class="mx-1" />
            <UButton size="md">示例按钮 MD</UButton>
            <span class="mx-1" />
            <UButton size="lg">示例按钮 LG</UButton>
          </div>
        </div>
      </div>
    </PanelBlock>

    <div v-if="store.loading" class="empty-state">
      正在加载项目设置...
    </div>

    <template v-else>
      <form class="section-grid" @submit.prevent="submit">
        <PanelBlock title="核心运行">
          <div class="grid gap-3 md:grid-cols-2">
            <FieldLabel label="PROXY_HOST">
              <input v-model="store.form.proxy_host" class="field-input" placeholder="127.0.0.1" />
            </FieldLabel>
            <FieldLabel label="PROXY_PORT">
              <input v-model="store.form.proxy_port" type="number" min="1" class="field-input" />
            </FieldLabel>
            <FieldLabel label="PROXY_READ_TIMEOUT_SECONDS">
              <input v-model="store.form.proxy_read_timeout_seconds" type="number" min="1" class="field-input" />
            </FieldLabel>
            <FieldLabel label="PROXY_WRITE_TIMEOUT_SECONDS">
              <input v-model="store.form.proxy_write_timeout_seconds" type="number" min="1" class="field-input" />
            </FieldLabel>
            <FieldLabel label="PROXY_SHUTDOWN_TIMEOUT_SECONDS">
              <input v-model="store.form.proxy_shutdown_timeout_seconds" type="number" min="1" class="field-input" />
            </FieldLabel>
          </div>
        </PanelBlock>

        <PanelBlock title="日志参数">
          <div class="grid gap-3 md:grid-cols-2">
            <FieldLabel label="PROXY_CHAIN_LOG_PATH">
              <input v-model="store.form.proxy_chain_log_path" class="field-input" placeholder=".data/icoo_proxy-chain.log" />
            </FieldLabel>
            <FieldLabel label="PROXY_CHAIN_LOG_MAX_BODY_BYTES">
              <input v-model="store.form.proxy_chain_log_max_body_bytes" type="number" min="0" class="field-input" />
            </FieldLabel>
          </div>
          <div class="mt-3">
            <label class="field-toggle">
              <input v-model="store.form.proxy_chain_log_bodies" type="checkbox" class="field-checkbox" />
              记录请求与响应体
            </label>
          </div>
        </PanelBlock>
      </form>
    </template>
  </section>
</template>

<script setup>
import { onMounted } from "vue";
import FieldLabel from "../components/FieldLabel.vue";
import PanelBlock from "../components/PanelBlock.vue";
import StatCard from "../components/StatCard.vue";
import UButton from "../components/ued/UButton.vue";
import { message } from "../components/ued/message";
import { useSettingsStore } from "../stores/settings";
import { useUiPrefsStore } from "../stores/uiPrefs";

const store = useSettingsStore();
const uiPrefs = useUiPrefsStore();

async function submit() {
  await store.save();
  if (!store.error) {
    message.success("项目设置已保存并重载。");
  }
}

onMounted(() => {
  store.load();
});
</script>
