<template>
  <section class="page-section">
    <Teleport to="#app-topbar-actions">
      <div class="app-topbar-actions__group">
        <button class="btn btn-secondary" @click="openAliasCreate">新建模型别名</button>
        <button class="btn btn-primary" @click="openSupplierCreate">新建供应商</button>
      </div>
    </Teleport>

    <div v-if="store.error" class="notice-error">
      {{ store.error }}
    </div>

    <div class="section-grid xl:grid-cols-5">
      <StatCard label="供应商总数" :value="String(store.items.length)" />
      <StatCard label="已启用供应商" :value="String(store.enabledCount)" />
      <StatCard label="已健康检查" :value="String(store.checkedCount)" />
      <StatCard label="已配置协议" :value="String(store.configuredPolicyCount)" />
      <StatCard label="启用中别名" :value="String(store.enabledAliasCount)" />
    </div>

    <div class="section-grid">
      <PanelBlock title="模型别名">
        <div class="mb-3 flex items-center justify-between gap-3">
          <div>
            <p class="text-sm font-medium text-slate-900">别名映射</p>
            <p class="mt-1 text-xs text-slate-500">为常用模型名称配置固定上游协议与目标模型。</p>
          </div>
          <UTag variant="info">总数：{{ store.aliases.length }}</UTag>
        </div>

        <div v-if="!store.aliases.length" class="empty-state">
          当前尚未配置模型别名。
        </div>
        <div v-else class="divide-y divide-[#eeeeF2] rounded-lg border border-[#e8e8ee]">
          <article
            v-for="item in store.aliases"
            :key="item.id"
            class="grid gap-3 px-3 py-3 lg:grid-cols-[1.4fr_2fr_auto] lg:items-center"
          >
            <div>
              <div class="flex items-center gap-2">
                <p class="text-base font-medium text-slate-900">{{ item.name }}</p>
                <UTag :variant="item.enabled ? 'success' : 'error'">
                  {{ item.enabled ? "启用" : "停用" }}
                </UTag>
              </div>
              <p class="mt-1 table-meta">更新时间：{{ formatCheckedAt(item.updated_at) }}</p>
            </div>
            <div class="grid gap-2 md:grid-cols-2">
              <div>
                <p class="table-meta">上游协议</p>
                <p class="mt-1 text-sm font-medium text-slate-900">{{ item.upstream_protocol }}</p>
              </div>
              <div>
                <p class="table-meta">目标模型</p>
                <p class="mt-1 break-all text-sm text-slate-700">{{ item.model }}</p>
              </div>
            </div>
            <div class="table-actions">
              <button class="btn btn-secondary" @click="openAliasEdit(item)">编辑</button>
              <button
                class="btn btn-error"
                :class="{ 'is-loading': store.aliasDeleting === item.id }"
                :disabled="store.aliasDeleting === item.id"
                @click="store.removeAlias(item.id)"
              >
                <span v-if="store.aliasDeleting === item.id" class="btn__spinner" />
                {{ store.aliasDeleting === item.id ? "删除中..." : "删除" }}
              </button>
            </div>
          </article>
        </div>
      </PanelBlock>
    </div>

    <div class="section-grid">
      <PanelBlock title="上下游管理">
        <div class="mb-3 flex items-center justify-between gap-3">
          <div>
            <p class="text-sm font-medium text-slate-900">协议映射</p>
            <p class="mt-1 text-xs text-slate-500">为三个下游协议分别指定供应商。默认模型直接继承供应商配置，不再单独设置目标模型。</p>
          </div>
          <UTag variant="info">启用中：{{ store.enabledPolicyCount }}</UTag>
        </div>

        <div class="divide-y divide-[#eeeeF2] rounded-lg border border-[#e8e8ee]">
          <article
            v-for="item in store.routeManagementRows"
            :key="item.key"
            class="grid gap-3 px-3 py-3 lg:grid-cols-[1.2fr_2fr_auto] lg:items-center"
          >
            <div>
              <div class="flex items-center gap-2">
                <p class="text-base font-medium text-slate-900">{{ item.label }}</p>
                <UTag code>{{ item.key }}</UTag>
              </div>
              <p class="mt-1 text-xs leading-5 text-slate-500">{{ item.description }}</p>
            </div>

            <div class="grid gap-2 md:grid-cols-3">
              <div>
                <p class="table-meta">供应商</p>
                <p class="mt-1 truncate text-sm font-medium text-slate-900">{{ item.supplierName }}</p>
              </div>
              <div>
                <p class="table-meta">上游协议</p>
                <p class="mt-1 truncate text-sm text-slate-700">{{ item.upstreamProtocol }}</p>
              </div>
              <div>
                <p class="table-meta">状态</p>
                <div class="mt-1">
                  <UTag :variant="item.statusVariant">{{ item.statusText }}</UTag>
                </div>
                <p class="mt-1 text-xs text-slate-500">{{ item.helperText }}</p>
              </div>
            </div>

            <div class="flex justify-end">
              <button
                class="btn btn-secondary"
                @click="item.policy ? openPolicyEdit(item.policy) : openPolicyCreate(item.key)"
              >
                {{ item.policy ? "编辑映射" : "配置映射" }}
              </button>
            </div>
          </article>
        </div>
      </PanelBlock>
    </div>

    <div class="section-grid">
      <PanelBlock title="供应商列表">
        <div v-if="store.loading" class="empty-state">
          正在加载供应商...
        </div>
        <div v-else-if="!store.items.length" class="empty-state">
          当前尚未配置供应商。
        </div>
        <UTable
          v-else
          :columns="supplierTableColumns"
          :rows="store.items"
          action-width="320px"
          fixed
          min-width="1480px"
          table-class="supplier-table"
        >
          <template #cell-supplier="{ row }">
            <div class="flex items-center gap-2">
              <p class="font-medium text-slate-900">{{ row.name }}</p>
              <UTag :variant="row.enabled ? 'success' : 'error'">
                {{ row.enabled ? "启用" : "停用" }}
              </UTag>
            </div>
            <p class="mt-1 text-sm leading-5 text-slate-600 table-cell-wrap">
              {{ row.description || "暂无描述。" }}
            </p>
            <p class="mt-1 table-meta">更新时间：{{ formatCheckedAt(row.updated_at) }}</p>
          </template>
          <template #cell-protocol="{ row }">
            <p class="font-medium text-slate-900">{{ row.protocol }}</p>
            <p class="mt-1 break-all table-meta table-cell-wrap">{{ row.base_url }}</p>
            <div class="mt-1 flex flex-wrap gap-1.5">
              <UTag code>{{ row.api_key_masked || "未保存 API Key" }}</UTag>
              <UTag v-if="row.only_stream" variant="warning">only_stream</UTag>
            </div>
            <p v-if="row.user_agent" class="mt-1 table-meta table-cell-wrap">UA: {{ row.user_agent }}</p>
          </template>
          <template #cell-models="{ row }">
            <div class="flex flex-wrap gap-2">
              <UTag
                v-for="model in row.models || []"
                :key="model"
                :variant="row.default_model === model ? 'success' : 'info'"
              >
                {{ row.default_model === model ? `${model} · 默认` : model }}
              </UTag>
              <span v-if="!(row.models || []).length" class="table-meta">无模型</span>
            </div>
            <p v-if="row.default_model && !(row.models || []).includes(row.default_model)" class="mt-1 text-xs text-amber-600">
              默认模型：{{ row.default_model }}
            </p>
          </template>
          <template #cell-health="{ row }">
            <template v-if="store.healthFor(row.id)">
              <div class="flex flex-wrap items-center gap-2">
                <UTag :variant="healthTone(store.healthFor(row.id))">
                  {{ store.healthFor(row.id).status }}
                </UTag>
                <UTag variant="info">{{ store.healthFor(row.id).duration_ms }} ms</UTag>
              </div>
              <p class="mt-1 table-meta">
                HTTP {{ store.healthFor(row.id).status_code || "无状态码" }}
              </p>
              <p class="mt-1 text-sm leading-5 text-slate-600 table-cell-wrap">
                {{ store.healthFor(row.id).message }}
              </p>
              <p class="mt-1 table-meta">{{ formatCheckedAt(store.healthFor(row.id).checked_at) }}</p>
            </template>
            <span v-else class="table-meta">尚未检查</span>
          </template>
          <template #actions="{ row }">
            <div class="table-actions">
              <button
                class="btn btn-info"
                :class="{ 'is-loading': store.checking === row.id }"
                :disabled="store.checking === row.id"
                @click="store.check(row.id)"
              >
                <span v-if="store.checking === row.id" class="btn__spinner" />
                {{ store.checking === row.id ? "检查中..." : "检查" }}
              </button>
              <button class="btn btn-secondary" @click="openSupplierEdit(row)">编辑</button>
              <button class="btn btn-secondary" @click="openModelEditor(row)">管理模型</button>
              <button
                class="btn btn-error"
                :class="{ 'is-loading': store.deleting === row.id }"
                :disabled="store.deleting === row.id"
                @click="openDeleteConfirm(row)"
              >
                <span v-if="store.deleting === row.id" class="btn__spinner" />
                {{ store.deleting === row.id ? "删除中..." : "删除" }}
              </button>
            </div>
          </template>
        </UTable>
      </PanelBlock>
    </div>

    <UModal
      v-model:open="supplierModalOpen"
      :title="store.form.id ? '编辑供应商' : '新建供应商'"
      width="640px"
      @close="store.resetForm"
    >
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
          <USelect
            v-model="store.form.default_model"
            label="默认模型"
            placeholder="请先在模型管理中配置候选模型"
            :options="supplierFormDefaultOptions"
          />
          <div class="rounded-lg border border-dashed border-[#d9dbe7] bg-slate-50 px-3 py-2 text-xs leading-5 text-slate-500">
            模型列表已拆分到独立弹窗管理。新建供应商后，可在列表中点击“管理模型”补充候选模型与默认模型。
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
          <button
            form="supplier-form"
            class="btn btn-primary"
            :class="{ 'is-loading': store.saving }"
            :disabled="store.saving"
          >
            <span v-if="store.saving" class="btn__spinner" />
            {{ store.saving ? "保存中..." : store.form.id ? "更新供应商" : "创建供应商" }}
          </button>
        </div>
      </template>
    </UModal>

    <UModal
      v-model:open="modelModalOpen"
      :title="store.modelForm.id ? `管理模型 · ${store.modelForm.name}` : '管理模型'"
      width="680px"
      @close="store.resetModelForm"
    >
      <form id="model-form" class="space-y-3" @submit.prevent="submitModelEditor">
        <div class="flex items-center justify-between gap-3">
          <div>
            <p class="text-sm font-medium text-slate-900">候选模型列表</p>
            <p class="mt-1 text-xs text-slate-500">默认模型必须来自当前候选模型列表。</p>
          </div>
          <button type="button" class="btn btn-secondary px-2 py-1 text-xs" @click="addModelRow(store.modelForm.models)">
            添加模型
          </button>
        </div>

        <div class="space-y-2">
          <div v-for="(model, index) in store.modelForm.models" :key="index" class="flex items-center gap-2">
            <input
              :value="model"
              class="field-input"
              :placeholder="index === 0 ? '例如：gpt-4.1-mini' : '继续添加模型'"
              @input="updateModelRow(store.modelForm.models, index, $event.target.value)"
            />
            <button
              type="button"
              class="btn btn-secondary shrink-0 px-2 py-2"
              :disabled="store.modelForm.models.length === 1"
              @click="removeModelRow(store.modelForm.models, index)"
            >
              删除
            </button>
          </div>
        </div>

        <USelect
          v-model="store.modelForm.default_model"
          label="默认模型"
          placeholder="不设置默认模型"
          :options="modelEditorDefaultOptions"
        />
      </form>
      <template #footer>
        <div class="flex justify-end gap-2">
          <button type="button" class="btn btn-secondary" @click="closeModelModal">取消</button>
          <button
            form="model-form"
            class="btn btn-primary"
            :class="{ 'is-loading': store.saving }"
            :disabled="store.saving"
          >
            <span v-if="store.saving" class="btn__spinner" />
            {{ store.saving ? "保存中..." : "保存模型设置" }}
          </button>
        </div>
      </template>
    </UModal>

    <UModal
      v-model:open="policyModalOpen"
      :title="store.policyForm.id ? '编辑路由策略' : '新建路由策略'"
      width="560px"
      @close="store.resetPolicyForm"
    >
      <form id="policy-form" class="space-y-3" @submit.prevent="submitPolicy">
        <div class="grid gap-3 md:grid-cols-2">
          <USelect v-model="store.policyForm.downstream_protocol" label="下游协议" :options="store.policyOptions" />
          <USelect v-model="store.policyForm.supplier_id" label="供应商" placeholder="请选择供应商" :options="supplierOptions" />
        </div>

        <div class="rounded-lg border border-dashed border-[#d9dbe7] bg-slate-50 px-3 py-2 text-xs leading-5 text-slate-500">
          默认模型不再在路由策略中单独设置，运行时会直接继承供应商的默认模型。
        </div>

        <label class="field-toggle">
          <input v-model="store.policyForm.enabled" type="checkbox" class="field-checkbox" />
          启用该路由策略
        </label>
      </form>
      <template #footer>
        <div class="flex justify-end gap-2">
          <button type="button" class="btn btn-secondary" @click="closePolicyModal">取消</button>
          <button
            form="policy-form"
            class="btn btn-primary"
            :class="{ 'is-loading': store.saving }"
            :disabled="store.saving"
          >
            <span v-if="store.saving" class="btn__spinner" />
            {{ store.saving ? "保存中..." : "保存路由策略" }}
          </button>
        </div>
      </template>
    </UModal>

    <UModal
      v-model:open="aliasModalOpen"
      :title="store.aliasForm.id ? '编辑模型别名' : '新建模型别名'"
      width="560px"
      @close="store.resetAliasForm"
    >
      <form id="alias-form" class="space-y-3" @submit.prevent="submitAlias">
        <FieldLabel label="别名名称">
          <input v-model="store.aliasForm.name" class="field-input" placeholder="例如：fast-model" />
        </FieldLabel>

        <div class="grid gap-3 md:grid-cols-2">
          <USelect v-model="store.aliasForm.upstream_protocol" label="上游协议" :options="store.aliasProtocolOptions" />
          <FieldLabel label="目标模型">
            <input v-model="store.aliasForm.model" class="field-input" placeholder="例如：gpt-4.1-mini" />
          </FieldLabel>
        </div>

        <label class="field-toggle">
          <input v-model="store.aliasForm.enabled" type="checkbox" class="field-checkbox" />
          启用该模型别名
        </label>
      </form>
      <template #footer>
        <div class="flex justify-end gap-2">
          <button type="button" class="btn btn-secondary" @click="closeAliasModal">取消</button>
          <button
            form="alias-form"
            class="btn btn-primary"
            :class="{ 'is-loading': store.saving }"
            :disabled="store.saving"
          >
            <span v-if="store.saving" class="btn__spinner" />
            {{ store.saving ? "保存中..." : "保存模型别名" }}
          </button>
        </div>
      </template>
    </UModal>

    <UConfirmDialog
      v-model:open="confirmState.open"
      title="确认删除供应商"
      :message="confirmState.message"
      description="删除后将同时移除该供应商对应的本地健康检查记录，且路由策略可能需要重新调整。"
      confirm-text="确认删除"
      cancel-text="取消"
      :loading="Boolean(store.deleting)"
      danger
      @confirm="confirmDelete"
    />
  </section>
</template>

<script setup>
import { computed, onMounted, reactive, ref } from "vue";
import { useSuppliersStore } from "../stores/suppliers";

import FieldLabel from "../components/FieldLabel.vue";
import PanelBlock from "../components/PanelBlock.vue";
import StatCard from "../components/StatCard.vue";
import UConfirmDialog from "../components/ued/UConfirmDialog.vue";
import UModal from "../components/ued/UModal.vue";
import USelect from "../components/ued/USelect.vue";
import UTable from "../components/ued/UTable.vue";
import UTag from "../components/ued/UTag.vue";

const store = useSuppliersStore();
const supplierModalOpen = ref(false);
const modelModalOpen = ref(false);
const policyModalOpen = ref(false);
const aliasModalOpen = ref(false);
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

const supplierTableColumns = [
  { key: "supplier", title: "供应商", width: "220px" },
  { key: "protocol", title: "协议 / 地址", width: "520px" },
  { key: "models", title: "模型 / 默认模型", width: "320px" },
  { key: "health", title: "健康状态", width: "220px" },
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
  confirmState.message = `确定要删除供应商“${item.name}”吗？`;
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
  await store.save();
  if (!store.error) {
    supplierModalOpen.value = false;
  }
}

async function submitModelEditor() {
  await store.saveModelEditor();
  if (!store.error) {
    modelModalOpen.value = false;
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
  await store.savePolicy();
  if (!store.error) {
    policyModalOpen.value = false;
  }
}

function openAliasCreate() {
  store.resetAliasForm();
  aliasModalOpen.value = true;
}

function openAliasEdit(item) {
  store.selectAlias(item);
  aliasModalOpen.value = true;
}

function closeAliasModal() {
  aliasModalOpen.value = false;
  store.resetAliasForm();
}

async function submitAlias() {
  await store.saveAlias();
  if (!store.error) {
    aliasModalOpen.value = false;
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
  }
}

onMounted(() => {
  store.load();
});
</script>
