<template>
  <section class="page-section">
    <div class="page-header">
      <h2 class="page-title">端点管理</h2>
      <div class="toolbar">
        <button class="btn btn-primary" @click="openCreate">新增端点</button>
        <button class="btn btn-secondary" :disabled="store.reloading" @click="store.reloadProxy">
          {{ store.reloading ? "重载中..." : "重载代理生效" }}
        </button>
      </div>
    </div>

    <div v-if="store.error" class="notice-error">
      {{ store.error }}
    </div>

    <div class="section-grid xl:grid-cols-4">
      <StatCard label="端点总数" :value="String(store.items.length)" />
      <StatCard label="已启用端点" :value="String(store.enabledCount)" />
      <StatCard label="自定义端点" :value="String(store.customCount)" />
      <StatCard label="生效方式" value="保存后重载代理" />
    </div>

    <PanelBlock title="代理端点">
      <div v-if="store.loading" class="empty-state">
        正在加载端点...
      </div>
      <div v-else-if="!store.items.length" class="empty-state">
        当前尚未配置端点。
      </div>
      <UTable v-else :columns="tableColumns" :rows="store.items" action-width="132px" fixed>
        <template #cell-path="{ row }">
          <UTag code>{{ row.path }}</UTag>
        </template>
        <template #cell-protocol="{ row }">
          <UTag variant="info">{{ row.protocol }}</UTag>
        </template>
        <template #cell-description="{ row }">
          <p class="max-w-xl text-sm text-slate-600">{{ row.description || "-" }}</p>
          <p class="mt-1 table-meta">更新时间：{{ formatDateTime(row.updated_at) }}</p>
        </template>
        <template #cell-builtIn="{ row }">
          <UTag :variant="row.built_in ? 'neutral' : 'warning'">
            {{ row.built_in ? "内置" : "自定义" }}
          </UTag>
        </template>
        <template #cell-enabled="{ row }">
          <UTag :variant="row.enabled ? 'success' : 'error'">
            {{ row.enabled ? "启用" : "停用" }}
          </UTag>
        </template>
        <template #actions="{ row }">
          <div class="table-actions">
            <button class="btn btn-secondary" @click="openEdit(row)">编辑</button>
            <button
              class="btn btn-error"
              :disabled="row.built_in || store.deleting === row.id"
              @click="remove(row)"
            >
              {{ store.deleting === row.id ? "删除中..." : "删除" }}
            </button>
          </div>
        </template>
      </UTable>
    </PanelBlock>

    <UModal
      v-model:open="modalOpen"
      :title="store.form.id ? '编辑端点' : '新增端点'"
      width="560px"
      @close="store.resetForm"
    >
      <form id="endpoint-form" class="space-y-3" @submit.prevent="submit">
        <FieldLabel label="路径">
          <input v-model="store.form.path" class="field-input" placeholder="/custom/v1/chat/completions" />
        </FieldLabel>
        <USelect v-model="store.form.protocol" label="协议" :options="store.protocolOptions" />
        <FieldLabel label="说明">
          <textarea v-model="store.form.description" class="field-input min-h-20" placeholder="描述该端点用途" />
        </FieldLabel>
        <label class="field-toggle">
          <input v-model="store.form.enabled" type="checkbox" class="field-checkbox" />
          启用该端点
        </label>
      </form>
      <template #footer>
        <div class="flex justify-end gap-2">
          <button type="button" class="btn btn-secondary" @click="closeModal">取消</button>
          <button form="endpoint-form" class="btn btn-primary" :disabled="store.saving">
            {{ store.saving ? "保存中..." : "保存端点" }}
          </button>
        </div>
      </template>
    </UModal>
  </section>
</template>

<script setup>
import { onMounted, ref } from "vue";
import FieldLabel from "../components/FieldLabel.vue";
import PanelBlock from "../components/PanelBlock.vue";
import StatCard from "../components/StatCard.vue";
import UModal from "../components/ued/UModal.vue";
import USelect from "../components/ued/USelect.vue";
import UTable from "../components/ued/UTable.vue";
import UTag from "../components/ued/UTag.vue";
import { useEndpointsStore } from "../stores/endpoints";

const store = useEndpointsStore();
const modalOpen = ref(false);
const tableColumns = [
  { key: "path", title: "路径", width: "22%" },
  { key: "protocol", title: "协议", width: "16%" },
  { key: "description", title: "说明", width: "34%" },
  { key: "builtIn", title: "类型", width: "10%" },
  { key: "enabled", title: "状态", width: "10%" },
];

function openCreate() {
  store.resetForm();
  modalOpen.value = true;
}

function openEdit(item) {
  store.select(item);
  modalOpen.value = true;
}

function closeModal() {
  modalOpen.value = false;
  store.resetForm();
}

async function submit() {
  await store.save();
  if (!store.error) {
    modalOpen.value = false;
  }
}

async function remove(item) {
  await store.remove(item.id);
}

function formatDateTime(value) {
  if (!value) {
    return "-";
  }
  return new Date(value).toLocaleString();
}

onMounted(() => {
  store.load();
});
</script>
