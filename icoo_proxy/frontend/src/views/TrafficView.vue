<template>
  <section class="page-section">
    <Teleport to="#app-topbar-actions">
      <div class="app-topbar-actions__group">
        <button
          class="btn btn-primary"
          :class="{ 'is-loading': store.refreshing }"
          :disabled="store.refreshing"
          @click="store.refresh"
        >
          <span v-if="store.refreshing" class="btn__spinner" />
          {{ store.refreshing ? "刷新中..." : "刷新流量" }}
        </button>
        <label class="field-toggle rounded-md">
          <input :checked="store.autoRefresh" type="checkbox" class="field-checkbox" @change="store.toggleAutoRefresh" />
          自动刷新
        </label>
      </div>
    </Teleport>

    <div v-if="store.error" class="notice-error">
      {{ store.error }}
    </div>

    <div class="section-grid grid-cols-2 lg:grid-cols-4">
      <StatCard icon="activity" label="最近请求数" :value="String(store.requests.length)" tone="info" />
      <StatCard icon="check" label="成功请求" :value="String(store.successCount)" tone="success" />
      <StatCard icon="alert" label="错误请求" :value="String(store.errorCount)" tone="danger" />
      <StatCard icon="timer" label="平均耗时" :value="`${store.averageLatency} ms`" />
    </div>

    <div class="section-grid lg:grid-cols-[280px_minmax(0,1fr)]">
      <PanelBlock title="筛选条件">
        <div class="space-y-3">
          <USelect
            label="协议"
            :model-value="store.filter"
            :options="store.protocolOptions"
            @update:model-value="store.setFilter"
          />

          <div class="divide-y divide-[#f0f0f0]">
            <div class="flex items-center justify-between gap-3 py-2">
              <p class="table-meta">最近刷新</p>
              <p class="text-right text-sm font-medium text-[#262626]">{{ formatDateTime(store.lastUpdatedAt) }}</p>
            </div>
            <div class="flex items-center justify-between gap-3 py-2">
              <p class="table-meta">筛选结果</p>
              <p class="text-right text-sm font-medium text-[#262626]">{{ store.filteredRequests.length }} 条</p>
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
            <p class="font-medium text-[#262626] table-cell-wrap">{{ row.request_id }}</p>
          </template>
          <template #cell-route="{ row }">
            <p class="text-sm text-[#595959] table-cell-wrap">{{ row.downstream }}</p>
            <p class="mt-0.5 table-meta table-cell-wrap">{{ row.upstream || "-" }}</p>
          </template>
          <template #cell-model="{ row }">
            <UTag code size="xs">{{ row.model || "-" }}</UTag>
          </template>
          <template #cell-status="{ row }">
            <UTag :variant="row.status_code >= 400 ? 'error' : 'success'" size="xs">
              {{ row.status_code || "-" }}
            </UTag>
          </template>
          <template #cell-duration="{ row }">
            <UTag size="xs">{{ row.duration_ms }} ms</UTag>
          </template>
          <template #cell-createdAt="{ row }">
            <span class="table-meta">{{ formatDateTime(row.created_at) }}</span>
          </template>
          <template #cell-error="{ row }">
            <p v-if="row.error" class="text-sm text-[#cf1322] table-cell-wrap">{{ row.error }}</p>
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
