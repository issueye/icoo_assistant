<template>
  <section class="page-section">
    <div class="page-header">
      <h2 class="page-title">本地 AI 网关运行概览</h2>
      <div class="toolbar">
        <button class="btn btn-primary" :disabled="store.refreshing" @click="store.reloadProxy">
          {{ store.refreshing ? "重载中..." : "重载代理" }}
        </button>
        <span class="badge" :class="store.data?.running ? 'badge-success' : 'badge-danger'">
          {{ store.data?.running ? "运行中" : "已停止" }}
        </span>
      </div>
    </div>

    <div v-if="store.error" class="notice-error">
      {{ store.error }}
    </div>

    <div v-if="store.loading" class="empty-state">
      正在加载网关概览...
    </div>

    <template v-else>
      <div class="section-grid xl:grid-cols-4">
        <StatCard label="代理地址" :value="store.data?.proxy_url || '未运行'" />
        <StatCard label="监听地址" :value="store.data?.listen_addr || '-'" />
        <StatCard label="访问模式" :value="store.data?.auth_required ? '需要 API Key' : '本地信任模式'" />
        <StatCard label="版本" :value="store.data?.version || '-'" />
      </div>

      <div class="section-grid xl:grid-cols-4">
        <StatCard label="供应商数量" :value="String(store.supplierCount)" />
        <StatCard label="已启用供应商" :value="String(store.enabledSupplierCount)" />
        <StatCard label="已健康检查" :value="String(store.checkedSupplierCount)" />
        <StatCard label="启用中策略" :value="String(store.activePolicyCount)" />
      </div>

      <div class="section-grid xl:grid-cols-2">
        <PanelBlock title="上游就绪状态">
          <div class="grid gap-3 md:grid-cols-3">
            <article
              v-for="upstream in store.data?.upstreams || []"
              :key="upstream.protocol"
              class="list-card"
            >
              <p class="text-xs font-semibold tracking-[0.04em] text-slate-500">{{ upstream.protocol }}</p>
              <p class="mt-2 text-lg font-semibold text-slate-900">{{ upstream.configured ? "已配置" : "缺少密钥" }}</p>
              <p class="mt-2 break-all text-sm text-slate-500">{{ upstream.base_url || "-" }}</p>
            </article>
          </div>
          <div class="mt-4 flex flex-wrap gap-2">
            <span
              v-for="(value, key) in store.checks"
              :key="key"
              class="badge"
              :class="value ? 'badge-success' : 'badge-warning'"
            >
              {{ key }}: {{ value }}
            </span>
          </div>
        </PanelBlock>

        <PanelBlock title="支持的接口路径">
          <div class="flex flex-wrap gap-2">
            <code v-for="route in store.routes" :key="route" class="mono-chip">
              {{ route }}
            </code>
          </div>
        </PanelBlock>
      </div>

      <div class="section-grid xl:grid-cols-2">
        <PanelBlock title="供应商健康汇总">
          <div class="grid gap-3 md:grid-cols-3">
            <article class="list-card">
              <p class="text-xs font-semibold tracking-[0.04em] text-emerald-600">可达</p>
              <p class="mt-2 text-2xl font-semibold text-slate-900">{{ store.reachableSupplierCount }}</p>
              <p class="mt-2 text-sm text-slate-500">最近检查结果正常的供应商数量。</p>
            </article>
            <article class="list-card">
              <p class="text-xs font-semibold tracking-[0.04em] text-amber-600">关注</p>
              <p class="mt-2 text-2xl font-semibold text-slate-900">{{ store.warningSupplierCount }}</p>
              <p class="mt-2 text-sm text-slate-500">返回 warning 或 unreachable 的供应商数量。</p>
            </article>
            <article class="list-card">
              <p class="text-xs font-semibold tracking-[0.04em] text-sky-600">未检查</p>
              <p class="mt-2 text-2xl font-semibold text-slate-900">{{ Math.max(store.supplierCount - store.checkedSupplierCount, 0) }}</p>
              <p class="mt-2 text-sm text-slate-500">当前会话中尚未执行健康检查的供应商数量。</p>
            </article>
          </div>
        </PanelBlock>

        <PanelBlock title="风险观察列表">
          <div v-if="store.unhealthySuppliers.length === 0" class="empty-state">
            暂无异常供应商，请在供应商页面执行健康检查后查看。
          </div>
          <div v-else class="space-y-3">
            <article v-for="item in store.unhealthySuppliers" :key="item.supplier_id" class="list-card border-rose-200 bg-rose-50">
              <div class="flex flex-wrap items-center justify-between gap-3">
                <div>
                  <p class="text-sm font-medium text-slate-900">{{ item.supplier_name }}</p>
                  <p class="mt-1 text-xs text-slate-500">{{ item.protocol }} | {{ item.base_url }}</p>
                </div>
                <span class="badge badge-danger">{{ item.status }}</span>
              </div>
              <p class="mt-3 text-sm text-rose-700">{{ item.message }}</p>
              <p class="mt-2 text-xs text-slate-500">
                {{ item.status_code || "no-status" }} | {{ item.duration_ms }} ms | {{ formatDateTime(item.checked_at) }}
              </p>
            </article>
          </div>
        </PanelBlock>
      </div>

      <div class="section-grid xl:grid-cols-2">
        <PanelBlock title="默认路由">
          <RouteList :items="store.data?.defaults || []" empty-text="当前尚未配置默认路由。" />
        </PanelBlock>

        <PanelBlock title="模型别名">
          <RouteList :items="store.data?.aliases || []" empty-text="当前尚未配置模型别名。" />
        </PanelBlock>
      </div>

      <PanelBlock title="默认路由策略">
        <div class="mb-4 flex flex-wrap gap-2">
          <span class="badge badge-success">启用：{{ store.activePolicyCount }}</span>
          <span class="badge badge-warning">停用：{{ store.inactivePolicyCount }}</span>
        </div>
        <div v-if="!(store.data?.route_policies || []).length" class="empty-state">
          当前尚未配置路由策略。
        </div>
        <div v-else class="grid gap-3 lg:grid-cols-3">
          <article v-for="policy in store.data?.route_policies || []" :key="policy.id" class="list-card">
            <div class="flex items-center justify-between gap-3">
              <p class="text-sm font-medium text-slate-900">{{ policy.downstream_protocol }}</p>
              <span class="badge" :class="policy.enabled ? 'badge-success' : 'badge-danger'">
                {{ policy.enabled ? "启用中" : "已停用" }}
              </span>
            </div>
            <p class="mt-3 text-sm text-slate-700">{{ policy.supplier_name || "未分配供应商" }}</p>
            <p class="mt-1 text-xs text-slate-500">{{ policy.upstream_protocol || "-" }} | {{ policy.target_model || "-" }}</p>
          </article>
        </div>
      </PanelBlock>

      <PanelBlock title="最近请求">
        <div v-if="store.requests.length === 0" class="empty-state">
          暂无请求记录。
        </div>
        <div v-else class="space-y-3">
          <article v-for="request in store.requests" :key="request.request_id" class="table-row">
            <div>
              <p class="text-sm font-medium text-slate-900">{{ request.downstream }} -> {{ request.upstream || "-" }}</p>
              <p class="mt-1 text-xs text-slate-500">{{ request.request_id }} | {{ request.created_at }}</p>
            </div>
            <div class="flex flex-wrap items-center gap-2">
              <code class="mono-chip">{{ request.model || "-" }}</code>
              <span class="badge" :class="request.status_code >= 400 ? 'badge-danger' : 'badge-success'">
                {{ request.status_code }}
              </span>
              <span class="tag-chip">{{ request.duration_ms }} ms</span>
            </div>
            <p v-if="request.error" class="w-full text-sm text-rose-700">{{ request.error }}</p>
          </article>
        </div>
      </PanelBlock>

      <PanelBlock title="当前版本说明">
        <ul class="space-y-2 text-sm text-slate-600">
          <li v-for="note in store.data?.notes || []" :key="note">{{ note }}</li>
        </ul>
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

const store = useOverviewStore();

function formatDateTime(value) {
  if (!value) {
    return "未检查";
  }
  return new Date(value).toLocaleString();
}

onMounted(() => {
  store.load();
});
</script>
