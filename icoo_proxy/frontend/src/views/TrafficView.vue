<template>
  <section class="page-section">
    <div class="page-header">
      <h2 class="page-title">请求流量监控</h2>
      <div class="toolbar">
        <button class="btn btn-primary" :disabled="store.refreshing" @click="store.refresh">
          {{ store.refreshing ? "刷新中..." : "刷新流量" }}
        </button>
        <label class="field-toggle rounded-lg">
          <input :checked="store.autoRefresh" type="checkbox" class="field-checkbox" @change="store.toggleAutoRefresh" />
          每 6 秒自动刷新
        </label>
      </div>
    </div>

    <div v-if="store.error" class="notice-error">
      {{ store.error }}
    </div>

    <div class="section-grid xl:grid-cols-4">
      <StatCard label="最近请求数" :value="String(store.requests.length)" />
      <StatCard label="成功请求数" :value="String(store.successCount)" />
      <StatCard label="错误请求数" :value="String(store.errorCount)" />
      <StatCard label="平均耗时" :value="`${store.averageLatency} ms`" />
    </div>

    <div class="section-grid xl:grid-cols-[320px_minmax(0,1fr)]">
      <PanelBlock title="筛选条件">
        <div class="space-y-4">
          <USelect
            label="协议"
            :model-value="store.filter"
            :options="store.protocolOptions"
            @update:model-value="store.setFilter"
          />

          <div class="divide-y divide-[#eeeeF2] border-y border-[#eeeeF2]">
            <div class="flex items-center justify-between gap-3 py-2">
              <p class="text-sm text-slate-500">最近刷新时间</p>
              <p class="text-right text-sm font-medium text-slate-900">{{ formatDateTime(store.lastUpdatedAt) }}</p>
            </div>
            <div class="flex items-center justify-between gap-3 py-2">
              <p class="text-sm text-slate-500">当前筛选结果</p>
              <p class="text-right text-sm font-medium text-slate-900">{{ store.filteredRequests.length }} 条</p>
            </div>
          </div>
        </div>
      </PanelBlock>

      <PanelBlock title="最近请求明细">
        <div v-if="store.loading" class="empty-state">
          正在加载流量数据...
        </div>
        <div v-else-if="store.filteredRequests.length === 0" class="empty-state">
          当前没有匹配的请求记录。
        </div>
        <div v-else class="table-shell">
          <div class="table-scroll">
            <table class="admin-table">
              <thead>
                <tr>
                  <th>请求 ID</th>
                  <th>下游 / 上游</th>
                  <th>模型</th>
                  <th>状态码</th>
                  <th>耗时</th>
                  <th>创建时间</th>
                  <th>错误信息</th>
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
                    <span class="badge" :class="request.status_code >= 400 ? 'badge-error' : 'badge-success'">
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
                    <span v-else class="table-meta">无</span>
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
import USelect from "../components/ued/USelect.vue";

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
    return "暂无";
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
