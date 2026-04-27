<template>
  <section class="page-section">
    <Teleport to="#app-topbar-actions">
      <div class="app-topbar-actions__group">
        <button class="btn btn-primary" @click="openCreate">新建</button>
      </div>
    </Teleport>

    <div v-if="store.error" class="notice-error">
      {{ store.error }}
    </div>

    <div class="section-grid grid-cols-3">
      <StatCard icon="tag" label="别名总数" :value="String(store.items.length)" tone="info" />
      <StatCard icon="check" label="已启用别名" :value="String(store.enabledCount)" tone="success" />
      <StatCard icon="supplier" label="关联供应商" :value="String(store.supplierCount)" />
    </div>

    <div class="section-grid">
      <PanelBlock title="模型别名列表">
        <div class="mb-3 flex items-center justify-between gap-3">
          <div>
            <p class="text-sm font-medium text-[#262626]">别名映射</p>
            <p class="mt-0.5 text-[11px] text-[#8c8c8c]">为常用模型名称配置上游协议与目标模型，关联供应商后自动继承协议。</p>
          </div>
          <UTag variant="info" size="xs">总数：{{ store.items.length }}</UTag>
        </div>

        <div v-if="store.loading" class="empty-state">
          正在加载别名...
        </div>
        <div v-else-if="!store.items.length" class="empty-state">
          当前尚未配置模型别名。
        </div>
        <div v-else class="divide-y divide-[#f0f0f0] rounded-lg border border-[#f0f0f0]">
          <article
            v-for="item in store.items"
            :key="item.id"
            class="grid gap-3 px-3 py-2.5 lg:grid-cols-[1.4fr_2fr_auto] lg:items-center"
          >
            <div>
              <div class="flex items-center gap-2">
                <p class="text-sm font-medium text-[#262626]">{{ item.name }}</p>
                <UTag :variant="item.enabled ? 'success' : 'error'" size="xs">
                  {{ item.enabled ? "启用" : "停用" }}
                </UTag>
              </div>
              <p class="mt-0.5 table-meta">更新时间：{{ formatTime(item.updated_at) }}</p>
            </div>
            <div class="grid gap-2 md:grid-cols-3">
              <div>
                <p class="table-meta">关联供应商</p>
                <p class="mt-0.5 text-sm font-medium text-[#262626]">{{ item.supplier_name || "未关联" }}</p>
              </div>
              <div>
                <p class="table-meta">上游协议</p>
                <p class="mt-0.5 truncate text-sm text-[#595959]">{{ item.upstream_protocol || "—" }}</p>
              </div>
              <div>
                <p class="table-meta">目标模型</p>
                <p class="mt-0.5 break-all text-sm text-[#595959]">{{ item.model }}</p>
              </div>
            </div>
            <div class="table-actions">
              <UIconButton icon="edit" label="编辑别名" @click="openEdit(item)" />
              <UIconButton
                icon="delete"
                label="删除别名"
                variant="error"
                :loading="store.deleting === item.id"
                :disabled="store.deleting === item.id"
                @click="removeAlias(item.id)"
              />
            </div>
          </article>
        </div>
      </PanelBlock>
    </div>

    <UModal
      v-model:open="modalOpen"
      :title="store.form.id ? '编辑模型别名' : '新建模型别名'"
      width="560px"
      @close="store.resetForm"
    >
      <form id="alias-form" class="space-y-3" @submit.prevent="submitAlias">
        <FieldLabel label="别名名称">
          <input v-model="store.form.name" class="field-input" placeholder="例如：fast-model" />
        </FieldLabel>

        <USelect
          v-model="store.form.supplier_id"
          label="关联供应商"
          placeholder="请选择供应商"
          :options="store.supplierOptions"
        />

        <FieldLabel label="上游协议" hint="根据所选供应商自动确定">
          <input
            :value="store.selectedSupplier?.protocol || '请先选择供应商'"
            class="field-input"
            disabled
          />
        </FieldLabel>

        <USelect
          v-model="store.form.model"
          label="目标模型"
          placeholder="请先选择供应商"
          :options="store.modelOptions"
          :disabled="!store.form.supplier_id"
          :hint="store.form.supplier_id && !store.modelOptions.length ? '所选供应商暂无模型' : ''"
        />

        <label class="field-toggle">
          <input v-model="store.form.enabled" type="checkbox" class="field-checkbox" />
          启用该模型别名
        </label>
      </form>
      <template #footer>
        <div class="flex justify-end gap-2">
          <button type="button" class="btn btn-secondary" @click="closeModal">取消</button>
          <button
            form="alias-form"
            class="btn btn-primary"
            :class="{ 'is-loading': store.saving }"
            :disabled="store.saving"
          >
            <span v-if="store.saving" class="btn__spinner" />
            {{ store.saving ? "保存中..." : store.form.id ? "更新别名" : "创建别名" }}
          </button>
        </div>
      </template>
    </UModal>
  </section>
</template>

<script setup>
import { onMounted, ref } from "vue";
import { useModelAliasesStore } from "../stores/modelAliases";

import FieldLabel from "../components/FieldLabel.vue";
import PanelBlock from "../components/PanelBlock.vue";
import StatCard from "../components/StatCard.vue";
import UIconButton from "../components/ued/UIconButton.vue";
import UModal from "../components/ued/UModal.vue";
import USelect from "../components/ued/USelect.vue";
import UTag from "../components/ued/UTag.vue";
import { message } from "../components/ued/message";

const store = useModelAliasesStore();
const modalOpen = ref(false);

function formatTime(value) {
  if (!value) {
    return "—";
  }
  return new Date(value).toLocaleString();
}

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

async function submitAlias() {
  const isEdit = Boolean(store.form.id);
  await store.save();
  if (!store.error) {
    modalOpen.value = false;
    message.success(isEdit ? "模型别名已更新。" : "模型别名已新增。");
  }
}

async function removeAlias(id) {
  await store.remove(id);
  if (!store.error) {
    message.success("模型别名已删除。");
  }
}

onMounted(() => {
  store.load();
});
</script>
