<template>
  <section class="page-section">
    <Teleport to="#app-topbar-actions">
      <div class="app-topbar-actions__group">
        <button class="btn btn-primary" @click="openSupplierCreate">新建</button>
      </div>
    </Teleport>

    <div v-if="store.error" class="notice-error">
      {{ store.error }}
    </div>

    <div class="section-grid grid-cols-2 lg:grid-cols-4">
      <StatCard icon="server" label="供应商总数" :value="String(store.items.length)" tone="info" />
      <StatCard icon="check" label="已启用" :value="String(store.enabledCount)" tone="success" />
      <StatCard icon="heart-pulse" label="已健康检查" :value="String(store.checkedCount)" />
      <StatCard icon="layers" label="已配置协议" :value="String(store.configuredPolicyCount)" tone="info" />
    </div>

    <div class="section-grid">
      <UTable :columns="routeManagementColumns" :rows="store.routeManagementRows" row-key="key" action-width="74px"
        size="small" table-class="route-management-table">
        <template #cell-protocol="{ row }">
          <div class="route-map__protocol-main">
            <p class="route-map__name">{{ row.label }}</p>
            <UTag code size="xs">{{ row.key }}</UTag>
          </div>
          <p class="route-map__desc">{{ row.description }}</p>
        </template>

        <template #cell-supplier="{ row }">
          <p class="route-map__value">{{ row.supplierName }}</p>
        </template>

        <template #cell-upstream="{ row }">
          <UTag code size="xs">{{ row.upstreamProtocol }}</UTag>
        </template>

          <template #cell-status="{ row }">
            <UTag :variant="row.statusVariant" size="xs" dot>{{ row.statusText }}</UTag>
          </template>

        <template #actions="{ row }">
          <div class="table-actions">
            <UIconButton icon="edit" :label="row.policy ? `编辑 ${row.label} 映射` : `配置 ${row.label} 映射`"
              @click="row.policy ? openPolicyEdit(row.policy) : openPolicyCreate(row.key)" />
          </div>
        </template>
      </UTable>
    </div>

    <div class="section-grid">
      <PanelBlock title="供应商列表">
        <div v-if="store.loading" class="empty-state">
          正在加载供应商...
        </div>
        <div v-else-if="!store.items.length" class="empty-state">
          当前尚未配置供应商。
        </div>
        <UTable v-else :columns="supplierTableColumns" :rows="store.items" action-width="148px" fixed min-width="1480px"
          table-class="supplier-table">
          <template #cell-supplier="{ row }">
            <div class="flex items-center gap-2">
              <p class="font-medium text-[#262626]">{{ row.name }}</p>
            </div>
            <p class="mt-0.5 text-sm leading-5 text-[#595959] table-cell-wrap">
              {{ row.description || "暂无描述。" }}
            </p>
          </template>
          <template #cell-protocol="{ row }">
            <p class="font-medium text-[#262626]">{{ row.protocol }}</p>
            <div class="mt-1 flex flex-wrap gap-1.5">
              <UTag v-if="row.only_stream" variant="warning" size="xs">only_stream</UTag>
            </div>
          </template>
          <template #cell-address="{ row }">
            <p class="mt-0.5 break-all table-meta table-cell-wrap">{{ row.base_url }}</p>
          </template>
          <template #cell-user_agent="{ row }">
            <p v-if="row.user_agent" class="mt-0.5 table-meta table-cell-wrap">UA: {{ row.user_agent }}</p>
          </template>
          <template #cell-key="{ row }">
            <UTag code size="xs">{{ row.api_key_masked || "未保存 API Key" }}</UTag>
          </template>
          <template #cell-models="{ row }">
            <div class="flex flex-wrap gap-1.5">
              <UTag v-for="model in row.models || []" :key="model"
                :variant="row.default_model === model ? 'success' : 'info'" size="xs">
                {{ row.default_model === model ? `${model} · 默认` : model }}
              </UTag>
              <span v-if="!(row.models || []).length" class="table-meta">无模型</span>
            </div>
          </template>
          <template #cell-health="{ row }">
            <template v-if="store.healthFor(row.id)">
              <div class="flex flex-wrap items-center gap-1.5">
                <UTag :variant="healthTone(store.healthFor(row.id))" size="xs">
                  {{ store.healthFor(row.id).status }}
                </UTag>
                <UTag variant="info" size="xs">{{ store.healthFor(row.id).duration_ms }} ms</UTag>
              </div>
              <p class="mt-0.5 table-meta">
                HTTP {{ store.healthFor(row.id).status_code || "无状态码" }}
              </p>
              <p class="mt-0.5 text-sm leading-5 text-[#595959] table-cell-wrap">
                {{ store.healthFor(row.id).message }}
              </p>
            </template>
            <span v-else class="table-meta">尚未检查</span>
          </template>
          <template #cell-status="{ row }">
            <UTag :variant="row.enabled ? 'success' : 'error'" size="xs">
              {{ row.enabled ? "启用" : "停用" }}
            </UTag>
          </template>
          <template #actions="{ row }">
            <div class="table-actions">
              <UIconButton icon="inspect" label="检查供应商" variant="info" :loading="store.checking === row.id"
                :disabled="store.checking === row.id" @click="checkSupplier(row.id)" />
              <UIconButton icon="edit" label="编辑供应商" @click="openSupplierEdit(row)" />
              <UIconButton icon="models" label="管理模型" @click="openModelEditor(row)" />
              <UIconButton icon="delete" label="删除供应商" variant="error" :loading="store.deleting === row.id"
                :disabled="store.deleting === row.id" @click="openDeleteConfirm(row)" />
            </div>
          </template>
        </UTable>
      </PanelBlock>
    </div>

    <UModal v-model:open="supplierModalOpen" :title="store.form.id ? '编辑供应商' : '新建供应商'" width="640px"
      @close="store.resetForm">
      <form id="supplier-form" class="space-y-3" @submit.prevent="submitSupplier">
        <div class="grid gap-3 md:grid-cols-2">
          <FieldLabel label="名称">
            <input v-model="store.form.name" class="field-input" placeholder="例如：OpenAI 生产环境" />
          </FieldLabel>
          <USelect v-model="store.form.protocol" label="协议" :options="protocolOptions" />
        </div>

        <FieldLabel label="基础地址">
          <input v-model="store.form.base_url" class="field-input" placeholder="https://api.openai.com" />
        </FieldLabel>

        <FieldLabel label="API Key">
          <input v-model="store.form.api_key" class="field-input" placeholder="编辑时留空则保留已有密钥" />
        </FieldLabel>

        <FieldLabel label="User-Agent">
          <input v-model="store.form.user_agent" class="field-input" placeholder="留空则使用默认上游 UA" />
        </FieldLabel>

        <FieldLabel label="描述">
          <textarea v-model="store.form.description" class="field-input min-h-24" placeholder="填写该供应商配置的用途说明" />
        </FieldLabel>

        <div class="grid gap-3 md:grid-cols-2">
          <USelect v-model="store.form.default_model" label="默认模型" placeholder="请先在模型管理中配置候选模型"
            :options="supplierFormDefaultOptions" />
          <div
            class="rounded-md border border-dashed border-[#d9d9d9] bg-[#fafafa] px-3 py-2 text-[11px] leading-5 text-[#8c8c8c]">
            模型列表已拆分到独立弹窗管理。新建供应商后，可在列表中点击"管理模型"补充候选模型与默认模型。
          </div>
        </div>

        <div class="grid gap-3 md:grid-cols-2">
          <label class="field-toggle">
            <input v-model="store.form.enabled" type="checkbox" class="field-checkbox" />
            启用该供应商配置
          </label>
          <label class="field-toggle">
            <input v-model="store.form.only_stream" type="checkbox" class="field-checkbox" />
            仅流式上游
          </label>
        </div>
      </form>
      <template #footer>
        <div class="flex justify-end gap-2">
          <button type="button" class="btn btn-secondary" @click="closeSupplierModal">取消</button>
          <button form="supplier-form" class="btn btn-primary" :class="{ 'is-loading': store.saving }"
            :disabled="store.saving">
            <span v-if="store.saving" class="btn__spinner" />
            {{ store.saving ? "保存中..." : store.form.id ? "更新供应商" : "创建供应商" }}
          </button>
        </div>
      </template>
    </UModal>

    <UModal v-model:open="modelModalOpen" :title="store.modelForm.id ? `管理模型 · ${store.modelForm.name}` : '管理模型'"
      width="680px" @close="store.resetModelForm">
      <form id="model-form" class="space-y-3" @submit.prevent="submitModelEditor">
        <div class="flex items-center justify-between gap-3">
          <div>
            <p class="text-sm font-medium text-[#262626]">候选模型列表</p>
            <p class="mt-0.5 text-[11px] text-[#8c8c8c]">默认模型必须来自当前候选模型列表。</p>
          </div>
          <button type="button" class="btn btn-secondary px-2 py-1 text-xs"
            @click="addModelRow(store.modelForm.models)">
            添加模型
          </button>
        </div>

        <div class="space-y-2">
          <div v-for="(model, index) in store.modelForm.models" :key="index" class="flex items-center gap-2">
            <input :value="model" class="field-input" :placeholder="index === 0 ? '例如：gpt-4.1-mini' : '继续添加模型'"
              @input="updateModelRow(store.modelForm.models, index, $event.target.value)" />
            <button type="button" class="btn btn-secondary shrink-0 px-2 py-2"
              :disabled="store.modelForm.models.length === 1" @click="removeModelRow(store.modelForm.models, index)">
              删除
            </button>
          </div>
        </div>

        <USelect v-model="store.modelForm.default_model" label="默认模型" placeholder="不设置默认模型"
          :options="modelEditorDefaultOptions" />
      </form>
      <template #footer>
        <div class="flex justify-end gap-2">
          <button type="button" class="btn btn-secondary" @click="closeModelModal">取消</button>
          <button form="model-form" class="btn btn-primary" :class="{ 'is-loading': store.saving }"
            :disabled="store.saving">
            <span v-if="store.saving" class="btn__spinner" />
            {{ store.saving ? "保存中..." : "保存模型设置" }}
          </button>
        </div>
      </template>
    </UModal>

    <UModal v-model:open="policyModalOpen" :title="store.policyForm.id ? '编辑路由策略' : '新建路由策略'" width="560px"
      @close="store.resetPolicyForm">
      <form id="policy-form" class="space-y-3" @submit.prevent="submitPolicy">
        <div class="grid gap-3 md:grid-cols-2">
          <USelect v-model="store.policyForm.downstream_protocol" label="下游协议" :options="store.policyOptions" />
          <USelect v-model="store.policyForm.supplier_id" label="供应商" placeholder="请选择供应商" :options="supplierOptions" />
        </div>

        <label class="field-toggle">
          <input v-model="store.policyForm.enabled" type="checkbox" class="field-checkbox" />
          启用该路由策略
        </label>
      </form>
      <template #footer>
        <div class="flex justify-end gap-2">
          <button type="button" class="btn btn-secondary" @click="closePolicyModal">取消</button>
          <button form="policy-form" class="btn btn-primary" :class="{ 'is-loading': store.saving }"
            :disabled="store.saving">
            <span v-if="store.saving" class="btn__spinner" />
            {{ store.saving ? "保存中..." : "保存路由策略" }}
          </button>
        </div>
      </template>
    </UModal>

    <UConfirmDialog v-model:open="confirmState.open" title="确认删除供应商" :message="confirmState.message"
      description="删除后将同时移除该供应商对应的本地健康检查记录，且路由策略可能需要重新调整。" confirm-text="确认删除" cancel-text="取消"
      :loading="Boolean(store.deleting)" danger @confirm="confirmDelete" />
  </section>
</template>

<script setup>
import { computed, onMounted, reactive, ref } from "vue";
import { useSuppliersStore } from "../stores/suppliers";

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

const store = useSuppliersStore();
const supplierModalOpen = ref(false);
const modelModalOpen = ref(false);
const policyModalOpen = ref(false);
const confirmState = reactive({
  open: false,
  id: "",
  message: "",
});

const protocolOptions = [
  { label: "anthropic", value: "anthropic" },
  { label: "openai-chat", value: "openai-chat" },
  { label: "openai-responses", value: "openai-responses" },
];

const supplierOptions = computed(() =>
  store.items.map((supplier) => ({
    label: `${supplier.name} (${supplier.protocol})`,
    value: supplier.id,
  })),
);

const routeManagementColumns = [
  { key: "protocol", title: "下游协议", width: "40%" },
  { key: "supplier", title: "供应商", width: "20%" },
  { key: "upstream", title: "上游协议", width: "20%" },
  { key: "status", title: "状态", width: "12%" },
];

const supplierTableColumns = [
  { key: "supplier", title: "供应商", width: "10%" },
  { key: "protocol", title: "协议 / 地址", width: "20%" },
  { key: "address", title: "地址", width: "20%" },
  { key: "user_agent", title: "User-Agent", width: "20%" },
  { key: "health", title: "健康状态", width: "15%" },
  { key: "status", title: "状态", width: "5%" },
];

const supplierFormDefaultOptions = computed(() => [
  { label: "不设置默认模型", value: "" },
  ...store.form.models
    .map((item) => String(item).trim())
    .filter(Boolean)
    .map((item) => ({ label: item, value: item })),
]);

const modelEditorDefaultOptions = computed(() => [
  { label: "不设置默认模型", value: "" },
  ...store.modelForm.models
    .map((item) => String(item).trim())
    .filter(Boolean)
    .map((item) => ({ label: item, value: item })),
]);

function healthTone(record) {
  if (!record) {
    return "neutral";
  }
  if (record.status === "reachable") {
    return "success";
  }
  if (record.status === "warning") {
    return "warning";
  }
  return "error";
}

function formatCheckedAt(value) {
  if (!value) {
    return "尚未检查";
  }
  return new Date(value).toLocaleString();
}

function openDeleteConfirm(item) {
  confirmState.open = true;
  confirmState.id = item.id;
  confirmState.message = `确定要删除供应商"${item.name}"吗？`;
}

function openSupplierCreate() {
  store.resetForm();
  supplierModalOpen.value = true;
}

function openSupplierEdit(item) {
  store.select(item);
  supplierModalOpen.value = true;
}

function closeSupplierModal() {
  supplierModalOpen.value = false;
  store.resetForm();
}

function openModelEditor(item) {
  store.selectModelEditor(item);
  modelModalOpen.value = true;
}

function closeModelModal() {
  modelModalOpen.value = false;
  store.resetModelForm();
}

function addModelRow(target) {
  target.push("");
}

function updateModelRow(target, index, value) {
  target[index] = value;
}

function removeModelRow(target, index) {
  if (target.length === 1) {
    return;
  }
  target.splice(index, 1);
}

async function submitSupplier() {
  const isEdit = Boolean(store.form.id);
  await store.save();
  if (!store.error) {
    supplierModalOpen.value = false;
    message.success(isEdit ? "供应商已更新。" : "供应商已新增。");
  }
}

async function submitModelEditor() {
  await store.saveModelEditor();
  if (!store.error) {
    modelModalOpen.value = false;
    message.success("模型设置已保存。");
  }
}

function openPolicyCreate(protocol = "anthropic") {
  store.resetPolicyForm();
  store.policyForm.downstream_protocol = protocol;
  policyModalOpen.value = true;
}

function openPolicyEdit(policy) {
  store.selectPolicy(policy);
  policyModalOpen.value = true;
}

function closePolicyModal() {
  policyModalOpen.value = false;
  store.resetPolicyForm();
}

async function submitPolicy() {
  const isEdit = Boolean(store.policyForm.id);
  await store.savePolicy();
  if (!store.error) {
    policyModalOpen.value = false;
    message.success(isEdit ? "路由策略已更新。" : "路由策略已新增。");
  }
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
    message.success("供应商已删除。");
  }
}

async function checkSupplier(id) {
  await store.check(id);
  if (!store.error) {
    message.success("供应商健康检查完成。");
  }
}

onMounted(() => {
  store.load();
});
</script>
