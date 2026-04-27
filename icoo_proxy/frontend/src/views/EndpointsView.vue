<template>
  <section class="page-section">
    <Teleport to="#app-topbar-actions">
      <div class="app-topbar-actions__group">
        <button class="btn btn-primary" @click="openCreate">新增端点</button>
        <button
          class="btn btn-secondary"
          :class="{ 'is-loading': store.reloading }"
          :disabled="store.reloading"
          @click="reloadProxy"
        >
          <span v-if="store.reloading" class="btn__spinner" />
          {{ store.reloading ? "重载中..." : "重载代理" }}
        </button>
      </div>
    </Teleport>

    <div v-if="store.error" class="notice-error">
      {{ store.error }}
    </div>

    <div class="section-grid grid-cols-2 lg:grid-cols-4">
      <StatCard icon="endpoint" label="端点总数" :value="String(store.items.length)" tone="info" />
      <StatCard icon="check" label="已启用" :value="String(store.enabledCount)" tone="success" />
      <StatCard icon="layers" label="自定义端点" :value="String(store.customCount)" />
    </div>

    <PanelBlock title="代理端点">
      <div v-if="store.loading" class="empty-state">
        正在加载端点...
      </div>
      <div v-else-if="!store.items.length" class="empty-state">
        当前尚未配置端点。
      </div>
      <UTable v-else :columns="tableColumns" :rows="store.items" action-width="90px" fixed>
        <template #cell-path="{ row }">
          <UTag code size="xs">{{ row.path }}</UTag>
        </template>
        <template #cell-protocol="{ row }">
          <UTag variant="info" size="xs">{{ row.protocol }}</UTag>
        </template>
        <template #cell-description="{ row }">
          <p class="max-w-xl text-sm text-[#595959]">{{ row.description || "-" }}</p>
          <p class="mt-0.5 table-meta">更新时间：{{ formatDateTime(row.updated_at) }}</p>
        </template>
        <template #cell-builtIn="{ row }">
          <UTag :variant="row.built_in ? 'neutral' : 'warning'" size="xs">
            {{ row.built_in ? "内置" : "自定义" }}
          </UTag>
        </template>
        <template #cell-enabled="{ row }">
          <UTag :variant="row.enabled ? 'success' : 'error'" size="xs">
            {{ row.enabled ? "启用" : "停用" }}
          </UTag>
        </template>
        <template #actions="{ row }">
          <div class="table-actions">
            <UIconButton icon="edit" label="编辑端点" @click="openEdit(row)" />
            <UIconButton
              icon="delete"
              label="删除端点"
              variant="error"
              :loading="store.deleting === row.id"
              :disabled="row.built_in || store.deleting === row.id"
              @click="openDeleteConfirm(row)"
            />
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
          <button
            form="endpoint-form"
            class="btn btn-primary"
            :class="{ 'is-loading': store.saving }"
            :disabled="store.saving"
          >
            <span v-if="store.saving" class="btn__spinner" />
            {{ store.saving ? "保存中..." : "保存端点" }}
          </button>
        </div>
      </template>
    </UModal>

    <UConfirmDialog
      v-model:open="confirmState.open"
      title="确认删除端点"
      :message="confirmState.message"
      description="删除后该端点路径将不再可用，保存后请重载代理生效。"
      confirm-text="确认删除"
      cancel-text="取消"
      :loading="Boolean(store.deleting)"
      danger
      @confirm="confirmDelete"
    />
  </section>
</template>

<script setup>
import { onMounted, reactive, ref } from "vue";
import FieldLabel from "../components/FieldLabel.vue";
import PanelBlock from "../components/PanelBlock.vue";
import StatCard from "../components/StatCard.vue";
import UConfirmDialog from "../components/ued/UConfirmDialog.vue";
import UIconButton from "../components/ued/UIconButton.vue";
import UModal from "../components/ued/UModal.vue";
import USelect from "../components/ued/USelect.vue";
import UTable from "../components/ued/UTable.vue";
import UTag from "../components/ued/UTag.vue";
import { message } from "../components/ued/message";
import { useEndpointsStore } from "../stores/endpoints";

const store = useEndpointsStore();
const modalOpen = ref(false);
const confirmState = reactive({
  open: false,
  id: "",
  message: "",
});
const tableColumns = [
  { key: "path", title: "路径", width: "22%" },
  { key: "protocol", title: "协议", width: "16%" },
  { key: "description", title: "说明", width: "65%" },
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
  const isEdit = Boolean(store.form.id);
  await store.save();
  if (!store.error) {
    modalOpen.value = false;
    message.success(isEdit ? "端点已更新。" : "端点已新增。");
  }
}

function openDeleteConfirm(item) {
  confirmState.open = true;
  confirmState.id = item.id;
  confirmState.message = `确定要删除端点"${item.path}"吗？`;
}

async function confirmDelete() {
  if (!confirmState.id) {
    return;
  }
  await store.remove(confirmState.id);
  if (!store.error) {
    confirmState.open = false;
    confirmState.id = "";
    confirmState.message = "";
    message.success("端点已删除。");
  }
}

async function reloadProxy() {
  await store.reloadProxy();
  if (!store.error) {
    message.success("代理已重载。");
  }
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
