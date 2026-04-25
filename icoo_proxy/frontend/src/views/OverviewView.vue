<template>
  <section class="page-section">
    <div class="page-header">
      <p class="page-eyebrow">Gateway Overview</p>
      <h2 class="page-title">本地 AI 网关运行概览</h2>
      <p class="page-description">
        这里集中展示网关运行状态、供应商健康情况、路由策略和最近请求，整体风格调整为更传统的后台管理界面，便于日常运维和交付使用。
      </p>
      <div class="toolbar">
        <button class="btn btn-primary" :disabled="store.refreshing" @click="store.reloadProxy">
          {{ store.refreshing ? "Reloading..." : "Reload Proxy" }}
        </button>
        <span class="badge" :class="store.data?.running ? 'badge-success' : 'badge-danger'">
          {{ store.data?.running ? "Running" : "Stopped" }}
        </span>
      </div>
    </div>

    <div v-if="store.error" class="notice-error">
      {{ store.error }}
    </div>

    <div v-if="store.loading" class="empty-state">
      Loading gateway overview...
    </div>

    <template v-else>
      <div class="section-grid xl:grid-cols-4">
        <StatCard label="Proxy URL" :value="store.data?.proxy_url || 'Not running'" />
        <StatCard label="Listen Address" :value="store.data?.listen_addr || '-'" />
        <StatCard label="Access Mode" :value="store.data?.auth_required ? 'API key required' : 'Trusted local mode'" />
        <StatCard label="Version" :value="store.data?.version || '-'" />
      </div>

      <div class="section-grid xl:grid-cols-4">
        <StatCard label="Suppliers" :value="String(store.supplierCount)" />
        <StatCard label="Enabled Suppliers" :value="String(store.enabledSupplierCount)" />
        <StatCard label="Health Checked" :value="String(store.checkedSupplierCount)" />
        <StatCard label="Active Policies" :value="String(store.activePolicyCount)" />
      </div>

      <div class="section-grid xl:grid-cols-2">
        <PanelBlock title="Provider Readiness" eyebrow="Runtime">
          <div class="grid gap-3 md:grid-cols-3">
            <article
              v-for="upstream in store.data?.upstreams || []"
              :key="upstream.protocol"
              class="list-card"
            >
              <p class="text-xs font-semibold uppercase tracking-[0.14em] text-slate-500">{{ upstream.protocol }}</p>
              <p class="mt-2 text-lg font-semibold text-slate-900">{{ upstream.configured ? "Configured" : "Missing key" }}</p>
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

        <PanelBlock title="Supported Endpoints" eyebrow="Surface Area">
          <div class="flex flex-wrap gap-2">
            <code v-for="route in store.routes" :key="route" class="mono-chip">
              {{ route }}
            </code>
          </div>
        </PanelBlock>
      </div>

      <div class="section-grid xl:grid-cols-2">
        <PanelBlock title="Supplier Health Summary" eyebrow="Observability">
          <div class="grid gap-3 md:grid-cols-3">
            <article class="list-card">
              <p class="text-xs font-semibold uppercase tracking-[0.14em] text-emerald-600">reachable</p>
              <p class="mt-2 text-2xl font-semibold text-slate-900">{{ store.reachableSupplierCount }}</p>
              <p class="mt-2 text-sm text-slate-500">最近检查结果正常的供应商数量。</p>
            </article>
            <article class="list-card">
              <p class="text-xs font-semibold uppercase tracking-[0.14em] text-amber-600">attention</p>
              <p class="mt-2 text-2xl font-semibold text-slate-900">{{ store.warningSupplierCount }}</p>
              <p class="mt-2 text-sm text-slate-500">返回 warning 或 unreachable 的供应商数量。</p>
            </article>
            <article class="list-card">
              <p class="text-xs font-semibold uppercase tracking-[0.14em] text-sky-600">unchecked</p>
              <p class="mt-2 text-2xl font-semibold text-slate-900">{{ Math.max(store.supplierCount - store.checkedSupplierCount, 0) }}</p>
              <p class="mt-2 text-sm text-slate-500">当前会话中尚未执行健康检查的供应商数量。</p>
            </article>
          </div>
        </PanelBlock>

        <PanelBlock title="Risk Watchlist" eyebrow="Needs Attention">
          <div v-if="store.unhealthySuppliers.length === 0" class="empty-state">
            No supplier warnings yet. Run checks from the supplier page to populate this watchlist.
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
        <PanelBlock title="Default Routes" eyebrow="Routing">
          <RouteList :items="store.data?.defaults || []" empty-text="No default routes configured yet." />
        </PanelBlock>

        <PanelBlock title="Model Aliases" eyebrow="Catalog">
          <RouteList :items="store.data?.aliases || []" empty-text="No aliases configured yet." />
        </PanelBlock>
      </div>

      <PanelBlock title="Default Route Policies" eyebrow="Supplier Binding">
        <div class="mb-4 flex flex-wrap gap-2">
          <span class="badge badge-success">active: {{ store.activePolicyCount }}</span>
          <span class="badge badge-warning">inactive: {{ store.inactivePolicyCount }}</span>
        </div>
        <div v-if="!(store.data?.route_policies || []).length" class="empty-state">
          No route policies configured yet.
        </div>
        <div v-else class="grid gap-3 lg:grid-cols-3">
          <article v-for="policy in store.data?.route_policies || []" :key="policy.id" class="list-card">
            <div class="flex items-center justify-between gap-3">
              <p class="text-sm font-medium text-slate-900">{{ policy.downstream_protocol }}</p>
              <span class="badge" :class="policy.enabled ? 'badge-success' : 'badge-danger'">
                {{ policy.enabled ? "active" : "inactive" }}
              </span>
            </div>
            <p class="mt-3 text-sm text-slate-700">{{ policy.supplier_name || "Unassigned supplier" }}</p>
            <p class="mt-1 text-xs text-slate-500">{{ policy.upstream_protocol || "-" }} | {{ policy.target_model || "-" }}</p>
          </article>
        </div>
      </PanelBlock>

      <PanelBlock title="Recent Requests" eyebrow="Traffic">
        <div v-if="store.requests.length === 0" class="empty-state">
          No requests recorded yet.
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

      <PanelBlock title="Current Build Notes" eyebrow="Scope">
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
    return "not checked";
  }
  return new Date(value).toLocaleString();
}

onMounted(() => {
  store.load();
});
</script>
