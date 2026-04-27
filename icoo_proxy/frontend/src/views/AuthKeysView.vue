<template>
  <section class="page-section">
    <Teleport to="#app-topbar-actions">
      <div class="app-topbar-actions__group">
        <button class="btn btn-primary" @click="openCreate">新增 Key</button>
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

    <div class="section-grid grid-cols-3">
      <StatCard icon="key" label="Key 总数" :value="String(store.items.length)" tone="info" />
      <StatCard icon="check" label="已启用" :value="String(store.enabledCount)" tone="success" />
      <StatCard icon="endpoint" label="使用方式" value="Bearer / x-api-key" />
    </div>

    <PanelBlock title="代理访问授权">
      <div v-if="store.loading" class="empty-state">
        正在加载授权 Key...
      </div>
      <div v-else-if="!store.items.length" class="empty-state">
        当前尚未添加授权 Key。本地信任模式仍按配置生效。
      </div>
      <UTable v-else :columns="tableColumns" :rows="store.items" action-width="118px" fixed>
        <template #cell-name="{ row }">
          <p class="font-medium text-[#262626]">{{ row.name }}</p>
          <p class="mt-0.5 table-meta">更新时间：{{ formatDateTime(row.updated_at) }}</p>
        </template>
        <template #cell-secret="{ row }">
          <UTag code size="xs">{{ row.secret_masked }}</UTag>
        </template>
        <template #cell-description="{ row }">
          <p class="max-w-xl text-sm text-[#595959]">{{ row.description || "-" }}</p>
        </template>
        <template #cell-enabled="{ row }">
          <UTag :variant="row.enabled ? 'success' : 'error'" size="xs">
            {{ row.enabled ? "启用" : "停用" }}
          </UTag>
        </template>
        <template #actions="{ row }">
          <div class="table-actions">
            <UIconButton
              icon="copy"
              label="复制授权 Key"
              variant="info"
              :loading="store.copying === row.id"
              :disabled="store.copying === row.id"
              @click="copyKey(row)"
            />
            <UIconButton icon="edit" label="编辑授权 Key" @click="openEdit(row)" />
            <UIconButton
              icon="delete"
              label="删除授权 Key"
              variant="error"
              :loading="store.deleting === row.id"
              :disabled="store.deleting === row.id"
              @click="openDeleteConfirm(row)"
            />
          </div>
        </template>
      </UTable>
    </PanelBlock>

    <UModal v-model:open="modalOpen" :title="store.form.id ? '编辑授权 Key' : '新增授权 Key'" width="560px" @close="store.resetForm">
      <form id="auth-key-form" class="space-y-3" @submit.prevent="submit">
        <FieldLabel label="名称">
          <input v-model="store.form.name" class="field-input" placeholder="本地开发 Key" />
        </FieldLabel>
        <FieldLabel label="Key">
          <div class="field-row">
            <input v-model="store.form.secret" class="field-input" :placeholder="store.form.id ? '留空则保留原 Key' : '输入或生成授权 Key'" />
            <button type="button" class="btn btn-secondary shrink-0" @click="generateSecret">生成</button>
          </div>
        </FieldLabel>
        <FieldLabel label="说明">
          <textarea v-model="store.form.description" class="field-input min-h-20" placeholder="描述该 Key 的使用方或用途" />
        </FieldLabel>
        <label class="field-toggle">
          <input v-model="store.form.enabled" type="checkbox" class="field-checkbox" />
          启用该 Key
        </label>
      </form>
      <template #footer>
        <div class="flex justify-end gap-2">
          <button type="button" class="btn btn-secondary" @click="closeModal">取消</button>
          <button
            form="auth-key-form"
            class="btn btn-primary"
            :class="{ 'is-loading': store.saving }"
            :disabled="store.saving"
          >
            <span v-if="store.saving" class="btn__spinner" />
            {{ store.saving ? "保存中..." : "保存 Key" }}
          </button>
        </div>
      </template>
    </UModal>

    <UConfirmDialog
      v-model:open="confirmState.open"
      title="确认删除授权 Key"
      :message="confirmState.message"
      description="删除后该 Key 将无法继续访问本地代理。"
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
import UTable from "../components/ued/UTable.vue";
import UTag from "../components/ued/UTag.vue";
import { message } from "../components/ued/message";
import { useAuthKeysStore } from "../stores/authKeys";

const store = useAuthKeysStore();
const modalOpen = ref(false);
const confirmState = reactive({
  open: false,
  id: "",
  message: "",
});
const tableColumns = [
  { key: "name", title: "名称", width: "22%" },
  { key: "secret", title: "Key", width: "22%" },
  { key: "description", title: "说明", width: "36%" },
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
    message.success(isEdit ? "授权 Key 已更新。" : "授权 Key 已新增。");
  }
}

function openDeleteConfirm(item) {
  confirmState.open = true;
  confirmState.id = item.id;
  confirmState.message = `确定要删除授权 Key"${item.name}"吗？`;
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
    message.success("授权 Key 已删除。");
  }
}

async function copyKey(item) {
  const secret = await store.copySecret(item.id);
  if (secret) {
    message.success("授权 Key 已复制。");
  }
}

function generateSecret() {
  store.generateSecret();
  message.success("授权 Key 已生成。");
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
