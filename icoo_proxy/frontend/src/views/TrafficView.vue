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
        <UTable
          v-else
          :columns="tableColumns"
          :rows="store.filteredRequests"
          row-key="request_id"
          fixed
          table-class="traffic-table"
        >
          <template #cell-requestId="{ row }">
            <p class="font-medium text-slate-900 table-cell-wrap">{{ row.request_id }}</p>
          </template>
          <template #cell-route="{ row }">
            <p class="text-sm text-slate-700 table-cell-wrap">{{ row.downstream }}</p>
            <p class="mt-1 table-meta table-cell-wrap">{{ row.upstream || "-" }}</p>
          </template>
          <template #cell-model="{ row }">
            <UTag code>{{ row.model || "-" }}</UTag>
          </template>
          <template #cell-status="{ row }">
            <UTag :variant="row.status_code >= 400 ? 'error' : 'success'">
              {{ row.status_code || "-" }}
            </UTag>
          </template>
          <template #cell-duration="{ row }">
            <UTag>{{ row.duration_ms }} ms</UTag>
          </template>
          <template #cell-createdAt="{ row }">
            <span class="table-meta">{{ formatDateTime(row.created_at) }}</span>
          </template>
          <template #cell-error="{ row }">
            <p v-if="row.error" class="text-sm text-rose-700 table-cell-wrap">{{ row.error }}</p>
            <span v-else class="table-meta">无</span>
          </template>
        </UTable>
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
import UTable from "../components/ued/UTable.vue";
import UTag from "../components/ued/UTag.vue";

const store = useTrafficStore();
let refreshTimer = null;
const tableColumns = [
  { key: "requestId", title: "请求 ID", width: "18%" },
  { key: "route", title: "下游 / 上游", width: "16%" },
  { key: "model", title: "模型", width: "15%" },
  { key: "status", title: "状态码", width: "10%" },
  { key: "duration", title: "耗时", width: "10%" },
  { key: "createdAt", title: "创建时间", width: "16%" },
  { key: "error", title: "错误信息", width: "15%" },
];

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
