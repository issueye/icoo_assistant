<template>
  <section class="page-section">
    <Teleport to="#app-topbar-actions">
      <div class="app-topbar-actions__group">
        <button
          class="btn btn-primary"
          :class="{ 'is-loading': store.refreshing }"
          :disabled="store.refreshing"
          @click="reloadProxy"
        >
          <span v-if="store.refreshing" class="btn__spinner" />
          {{ store.refreshing ? "重载中..." : "重载代理" }}
        </button>
        <span class="badge" :class="store.data?.running ? 'badge-success' : 'badge-error'">
          {{ store.data?.running ? "运行中" : "已停止" }}
        </span>
      </div>
    </Teleport>

    <div v-if="store.error" class="notice-error">
      {{ store.error }}
    </div>

    <div v-if="store.loading" class="empty-state">
      正在加载网关概览...
    </div>

    <template v-else>
      <div class="section-grid grid-cols-2 lg:grid-cols-4">
        <StatCard icon="wifi" label="监听地址" :value="store.data?.listen_addr || '-'" />
        <StatCard icon="key" label="访问模式" :value="store.data?.auth_required ? `${store.data?.auth_key_count || 0} 个授权 Key` : '本地信任模式'" :tone="store.data?.auth_required ? 'success' : 'warning'" />
        <StatCard icon="server" label="供应商" :value="String(store.supplierCount)" tone="info" />
        <StatCard icon="layers" label="启用策略" :value="String(store.activePolicyCount)" tone="info" />
      </div>

      <div class="section-grid lg:grid-cols-2">
        <PanelBlock title="上游就绪状态">
          <div class="divide-y divide-[#f0f0f0]">
            <div v-for="upstream in store.data?.upstreams || []" :key="upstream.protocol" class="grid gap-2 py-2.5 grid-cols-[1fr_auto] items-center">
              <div class="min-w-0">
                <p class="text-sm font-medium text-[#262626]">{{ upstream.protocol }}</p>
                <p class="mt-0.5 truncate text-xs text-[#8c8c8c]">{{ upstream.base_url || "-" }}</p>
              </div>
              <UTag :variant="upstream.configured ? 'success' : 'warning'" size="xs">
                {{ upstream.configured ? "已配置" : "缺少密钥" }}
              </UTag>
            </div>
          </div>
          <div class="mt-3 flex flex-wrap gap-1.5">
            <UTag
              v-for="(value, key) in store.checks"
              :key="key"
              :variant="value ? 'success' : 'warning'"
              size="xs"
            >
              {{ key }}: {{ value }}
            </UTag>
          </div>
        </PanelBlock>

        <PanelBlock title="支持的接口路径">
          <div class="flex flex-wrap gap-1.5">
            <UTag v-for="route in store.routes" :key="route" code size="xs">
              {{ route }}
            </UTag>
          </div>
        </PanelBlock>
      </div>

      <div class="section-grid lg:grid-cols-2">
        <PanelBlock title="供应商健康汇总">
          <div class="divide-y divide-[#f0f0f0]">
            <div class="grid gap-2 py-2.5 grid-cols-[auto_1fr] items-center">
              <UTag variant="success" size="xs">可达</UTag>
              <div class="flex items-baseline gap-2">
                <p class="text-lg font-semibold text-[#262626]">{{ store.reachableSupplierCount }}</p>
                <p class="text-xs text-[#8c8c8c]">最近检查结果正常的供应商数量</p>
              </div>
            </div>
            <div class="grid gap-2 py-2.5 grid-cols-[auto_1fr] items-center">
              <UTag variant="warning" size="xs">关注</UTag>
              <div class="flex items-baseline gap-2">
                <p class="text-lg font-semibold text-[#262626]">{{ store.warningSupplierCount }}</p>
                <p class="text-xs text-[#8c8c8c]">返回 warning 或 unreachable 的供应商数量</p>
              </div>
            </div>
            <div class="grid gap-2 py-2.5 grid-cols-[auto_1fr] items-center">
              <UTag variant="info" size="xs">未检查</UTag>
              <div class="flex items-baseline gap-2">
                <p class="text-lg font-semibold text-[#262626]">{{ Math.max(store.supplierCount - store.checkedSupplierCount, 0) }}</p>
                <p class="text-xs text-[#8c8c8c]">尚未执行健康检查的供应商数量</p>
              </div>
            </div>
          </div>
        </PanelBlock>

        <PanelBlock title="默认路由">
          <RouteList :items="store.data?.defaults || []" empty-text="当前尚未配置默认路由。" />
        </PanelBlock>
      </div>

      <PanelBlock title="默认路由策略">
        <div class="mb-3 flex flex-wrap gap-1.5">
          <UTag variant="success" size="xs">启用：{{ store.activePolicyCount }}</UTag>
          <UTag variant="warning" size="xs">停用：{{ store.inactivePolicyCount }}</UTag>
        </div>
        <div v-if="!(store.data?.route_policies || []).length" class="empty-state">
          当前尚未配置路由策略。
        </div>
        <div v-else class="divide-y divide-[#f0f0f0]">
          <article v-for="policy in store.data?.route_policies || []" :key="policy.id" class="grid gap-3 py-2.5 lg:grid-cols-[1fr_1fr_1fr_auto] items-center">
            <div>
              <p class="text-sm font-medium text-[#262626]">{{ policy.downstream_protocol }}</p>
              <p class="mt-0.5 text-[11px] text-[#8c8c8c]">下游协议</p>
            </div>
            <p class="text-sm text-[#595959]">{{ policy.supplier_name || "未分配供应商" }}</p>
            <p class="text-xs text-[#8c8c8c]">{{ policy.upstream_protocol || "-" }} | 默认模型来自供应商</p>
            <UTag :variant="policy.enabled ? 'success' : 'error'" size="xs">
              {{ policy.enabled ? "启用中" : "已停用" }}
            </UTag>
          </article>
        </div>
      </PanelBlock>
    </template>
  </section>
</template>

<script setup>
import { onMounted } from "vue";
import { useOverviewStore } from "../stores/overview";

import PanelBlock from "../components/PanelBlock.vue";
import RouteList from "../components/RouteList.vue";
import StatCard from "../components/StatCard.vue";
import UTag from "../components/ued/UTag.vue";
import { message } from "../components/ued/message";

const store = useOverviewStore();

onMounted(() => {
  store.load();
});

async function reloadProxy() {
  await store.reloadProxy();
  if (!store.error) {
    message.success("代理已重载。");
  }
}
</script>
