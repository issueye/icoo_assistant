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
      <div v-else class="table-shell">
        <div class="table-scroll">
          <table class="admin-table">
            <thead>
              <tr>
                <th>路径</th>
                <th>协议</th>
                <th>说明</th>
                <th>类型</th>
                <th>状态</th>
                <th>操作</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="item in store.items" :key="item.id">
                <td>
                  <UTag code>{{ item.path }}</UTag>
                </td>
                <td>
                  <UTag variant="info">{{ item.protocol }}</UTag>
                </td>
                <td>
                  <p class="max-w-xl text-sm text-slate-600">{{ item.description || "-" }}</p>
                  <p class="mt-1 table-meta">更新时间：{{ formatDateTime(item.updated_at) }}</p>
                </td>
                <td>
                  <UTag :variant="item.built_in ? 'neutral' : 'warning'">
                    {{ item.built_in ? "内置" : "自定义" }}
                  </UTag>
                </td>
                <td>
                  <UTag :variant="item.enabled ? 'success' : 'error'">
                    {{ item.enabled ? "启用" : "停用" }}
                  </UTag>
                </td>
                <td>
                  <div class="table-actions">
                    <button class="btn btn-secondary" @click="openEdit(item)">编辑</button>
                    <button
                      class="btn btn-error"
                      :disabled="item.built_in || store.deleting === item.id"
                      @click="remove(item)"
                    >
                      {{ store.deleting === item.id ? "删除中..." : "删除" }}
                    </button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
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
import UTag from "../components/ued/UTag.vue";
import { useEndpointsStore } from "../stores/endpoints";

const store = useEndpointsStore();
const modalOpen = ref(false);

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
