<template>
  <section class="page-section">
    <div class="page-header">
      <p class="page-eyebrow">Traffic Monitor</p>
      <h2 class="page-title">请求流量监控</h2>
      <p class="page-description">
        用更接近传统后台系统的方式查看最近请求、状态码和耗时，便于定位协议转换异常、路由命中问题和上游响应波动。
      </p>
      <div class="toolbar">
        <button class="btn btn-primary" :disabled="store.refreshing" @click="store.refresh">
          {{ store.refreshing ? "Refreshing..." : "Refresh Traffic" }}
        </button>
        <label class="field-toggle rounded-lg">
          <input :checked="store.autoRefresh" type="checkbox" class="field-checkbox" @change="store.toggleAutoRefresh" />
          Auto refresh every 6s
        </label>
      </div>
    </div>

    <div v-if="store.error" class="notice-error">
      {{ store.error }}
    </div>

    <div class="section-grid xl:grid-cols-4">
      <StatCard label="Recent Requests" :value="String(store.requests.length)" />
      <StatCard label="Successful" :value="String(store.successCount)" />
      <StatCard label="Errors" :value="String(store.errorCount)" />
      <StatCard label="Avg Latency" :value="`${store.averageLatency} ms`" />
    </div>

    <div class="section-grid xl:grid-cols-[320px_minmax(0,1fr)]">
      <PanelBlock title="Filters" eyebrow="Focus">
        <div class="space-y-4">
          <label class="block">
            <span class="mb-2 block text-sm font-medium text-slate-700">Protocol</span>
            <select :value="store.filter" class="field-input" @change="store.setFilter($event.target.value)">
              <option v-for="option in store.protocolOptions" :key="option" :value="option">
                {{ option }}
              </option>
            </select>
          </label>

          <div class="sub-card">
            <p class="text-sm font-medium text-slate-900">Last updated</p>
            <p class="mt-2 text-sm text-slate-500">{{ formatDateTime(store.lastUpdatedAt) }}</p>
          </div>

          <div class="sub-card">
            <p class="text-sm font-medium text-slate-900">Current filter result</p>
            <p class="mt-2 text-sm text-slate-500">{{ store.filteredRequests.length }} request(s) visible</p>
          </div>
        </div>
      </PanelBlock>

      <PanelBlock title="Recent Request Timeline" eyebrow="Inspection">
        <div v-if="store.loading" class="empty-state">
          Loading traffic...
        </div>
        <div v-else-if="store.filteredRequests.length === 0" class="empty-state">
          No matching requests yet.
        </div>
        <div v-else class="table-shell">
          <div class="table-scroll">
            <table class="admin-table">
              <thead>
                <tr>
                  <th>Request ID</th>
                  <th>Downstream / Upstream</th>
                  <th>Model</th>
                  <th>Status</th>
                  <th>Latency</th>
                  <th>Created At</th>
                  <th>Error</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="request in store.filteredRequests" :key="request.request_id">
                  <td>
                    <p class="font-medium text-slate-900">{{ request.request_id }}</p>
                  </td>
                  <td>
                    <p class="text-sm text-slate-700">{{ request.downstream }}</p>
                    <p class="mt-1 table-meta">{{ request.upstream || "-" }}</p>
                  </td>
                  <td>
                    <code class="mono-chip">{{ request.model || "-" }}</code>
                  </td>
                  <td>
                    <span class="badge" :class="request.status_code >= 400 ? 'badge-danger' : 'badge-success'">
                      {{ request.status_code || "-" }}
                    </span>
                  </td>
                  <td>
                    <span class="tag-chip">{{ request.duration_ms }} ms</span>
                  </td>
                  <td>
                    <span class="table-meta">{{ formatDateTime(request.created_at) }}</span>
                  </td>
                  <td>
                    <p v-if="request.error" class="text-sm text-rose-700">{{ request.error }}</p>
                    <span v-else class="table-meta">-</span>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </PanelBlock>
    </div>
  </section>
</template>

<script setup>
import { onBeforeUnmount, onMounted, watch } from "vue";
import { useTrafficStore } from "../stores/traffic";

import PanelBlock from "../components/PanelBlock.vue";
import StatCard from "../components/StatCard.vue";

const store = useTrafficStore();
let refreshTimer = null;

function stopTimer() {
  if (refreshTimer) {
    clearInterval(refreshTimer);
    refreshTimer = null;
  }
}

function startTimer() {
  stopTimer();
  if (!store.autoRefresh) {
    return;
  }
  refreshTimer = setInterval(() => {
    store.refresh();
  }, 6000);
}

function formatDateTime(value) {
  if (!value) {
    return "not available";
  }
  return new Date(value).toLocaleString();
}

watch(
  () => store.autoRefresh,
  () => {
    startTimer();
  },
);

onMounted(() => {
  store.load();
  startTimer();
});

onBeforeUnmount(() => {
  stopTimer();
});
</script>
