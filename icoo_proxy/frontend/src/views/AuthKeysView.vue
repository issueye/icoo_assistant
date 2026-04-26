<template>
  <section class="page-section">
    <Teleport to="#app-topbar-actions">
      <div class="app-topbar-actions__group">
        <button class="btn btn-primary" @click="openCreate">新增 Key</button>
        <button class="btn btn-secondary" :disabled="store.reloading" @click="store.reloadProxy">
          {{ store.reloading ? "重载中..." : "重载代理生效" }}
        </button>
      </div>
    </Teleport>

    <div v-if="store.error" class="notice-error">
      {{ store.error }}
    </div>

    <div class="section-grid xl:grid-cols-3">
      <StatCard label="Key 总数" :value="String(store.items.length)" />
      <StatCard label="已启用" :value="String(store.enabledCount)" />
      <StatCard label="使用方式" value="Bearer / x-api-key" />
    </div>

    <PanelBlock title="代理访问授权">
      <div v-if="store.loading" class="empty-state">
        正在加载授权 Key...
      </div>
      <div v-else-if="!store.items.length" class="empty-state">
        当前尚未添加授权 Key。本地信任模式仍按配置生效。
      </div>
      <UTable v-else :columns="tableColumns" :rows="store.items" action-width="220px" fixed>
        <template #cell-name="{ row }">
          <p class="font-medium text-slate-950">{{ row.name }}</p>
          <p class="mt-1 table-meta">更新时间：{{ formatDateTime(row.updated_at) }}</p>
        </template>
        <template #cell-secret="{ row }">
          <UTag code>{{ row.secret_masked }}</UTag>
        </template>
        <template #cell-description="{ row }">
          <p class="max-w-xl text-sm text-slate-600">{{ row.description || "-" }}</p>
        </template>
        <template #cell-enabled="{ row }">
          <UTag :variant="row.enabled ? 'success' : 'error'">
            {{ row.enabled ? "启用" : "停用" }}
          </UTag>
        </template>
        <template #actions="{ row }">
          <div class="table-actions">
            <button
              class="btn btn-info"
              :disabled="store.copying === row.id"
              @click="copyKey(row)"
            >
              {{ store.copying === row.id ? "复制中..." : "复制" }}
            </button>
            <button class="btn btn-secondary" @click="openEdit(row)">编辑</button>
            <button class="btn btn-error" :disabled="store.deleting === row.id" @click="remove(row)">
              {{ store.deleting === row.id ? "删除中..." : "删除" }}
            </button>
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
            <button type="button" class="btn btn-secondary shrink-0" @click="store.generateSecret">生成</button>
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
          <button form="auth-key-form" class="btn btn-primary" :disabled="store.saving">
            {{ store.saving ? "保存中..." : "保存 Key" }}
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
import UTable from "../components/ued/UTable.vue";
import UTag from "../components/ued/UTag.vue";
import { useAuthKeysStore } from "../stores/authKeys";

const store = useAuthKeysStore();
const modalOpen = ref(false);
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
  await store.save();
  if (!store.error) {
    modalOpen.value = false;
  }
}

async function remove(item) {
  await store.remove(item.id);
}

async function copyKey(item) {
  const secret = await store.copySecret(item.id);
  if (secret) {
    // Optionally show a toast or alert; for now rely on clipboard feedback
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
