<template>
  <section class="page-section">
    <div class="page-header">
      <h2 class="page-title">供应商与路由策略管理</h2>
      <div class="toolbar">
        <button class="btn btn-primary" @click="openSupplierCreate">新建供应商</button>
        <button class="btn btn-secondary" @click="openPolicyCreate">新建路由策略</button>
      </div>
    </div>

    <div v-if="store.error" class="notice-error">
      {{ store.error }}
    </div>

    <div class="section-grid xl:grid-cols-4">
      <StatCard label="供应商总数" :value="String(store.items.length)" />
      <StatCard label="已启用配置" :value="String(store.enabledCount)" />
      <StatCard label="已健康检查" :value="String(store.checkedCount)" />
      <StatCard label="路由策略数" :value="String(store.policies.length)" />
    </div>

    <div class="section-grid">
      <PanelBlock title="供应商列表">
        <div v-if="store.loading" class="empty-state">
          正在加载供应商...
        </div>
        <div v-else-if="!store.items.length" class="empty-state">
          当前尚未配置供应商。
        </div>
        <div v-else class="table-shell">
          <div class="table-scroll">
            <table class="admin-table">
              <thead>
                <tr>
                  <th>供应商</th>
                  <th>协议 / 地址</th>
                  <th>模型 / 标签</th>
                  <th>健康状态</th>
                  <th>操作</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="item in store.items" :key="item.id">
                  <td>
                    <div class="flex items-center gap-2">
                      <p class="font-medium text-slate-900">{{ item.name }}</p>
                      <span class="badge" :class="item.enabled ? 'badge-success' : 'badge-danger'">
                        {{ item.enabled ? "启用" : "停用" }}
                      </span>
                    </div>
                    <p class="mt-2 text-sm text-slate-600">{{ item.description || "暂无描述。" }}</p>
                    <p class="mt-2 table-meta">更新时间：{{ formatCheckedAt(item.updated_at) }}</p>
                  </td>
                  <td>
                    <p class="font-medium text-slate-900">{{ item.protocol }}</p>
                    <p class="mt-2 break-all table-meta">{{ item.base_url }}</p>
                    <div class="mt-2">
                      <code class="mono-chip">{{ item.api_key_masked || "未保存 API Key" }}</code>
                    </div>
                  </td>
                  <td>
                    <div class="flex flex-wrap gap-2">
                      <span v-for="model in item.models || []" :key="model" class="tag-chip">
                        {{ model }}
                      </span>
                      <span v-if="!(item.models || []).length" class="table-meta">无模型</span>
                    </div>
                    <div class="mt-2 flex flex-wrap gap-2">
                      <span v-for="tag in item.tags || []" :key="tag" class="tag-chip">
                        #{{ tag }}
                      </span>
                    </div>
                  </td>
                  <td>
                    <template v-if="store.healthFor(item.id)">
                      <div class="flex flex-wrap items-center gap-2">
                        <span class="badge" :class="healthTone(store.healthFor(item.id)).badge">
                          {{ store.healthFor(item.id).status }}
                        </span>
                        <span class="tag-chip">{{ store.healthFor(item.id).duration_ms }} ms</span>
                      </div>
                      <p class="mt-2 table-meta">
                        HTTP {{ store.healthFor(item.id).status_code || "无状态码" }}
                      </p>
                      <p class="mt-2 text-sm text-slate-600">{{ store.healthFor(item.id).message }}</p>
                      <p class="mt-2 table-meta">{{ formatCheckedAt(store.healthFor(item.id).checked_at) }}</p>
                    </template>
                    <span v-else class="table-meta">尚未检查</span>
                  </td>
                  <td>
                    <div class="table-actions">
                      <button class="btn btn-info" :disabled="store.checking === item.id" @click="store.check(item.id)">
                        {{ store.checking === item.id ? "检查中..." : "检查" }}
                      </button>
                      <button class="btn btn-secondary" @click="openSupplierEdit(item)">编辑</button>
                      <button class="btn btn-danger" :disabled="store.deleting === item.id" @click="openDeleteConfirm(item)">
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
    </div>

    <div class="section-grid xl:grid-cols-2">
      <PanelBlock title="三条默认协议路由">
        <div class="grid gap-3">
          <article v-for="item in store.policiesByProtocol" :key="item.key" class="list-card">
            <div class="flex flex-wrap items-start justify-between gap-3">
              <div>
                <p class="text-base font-medium text-slate-900">{{ item.label }}</p>
                <p class="mt-2 text-sm text-slate-600">{{ item.description }}</p>
              </div>
              <span
                class="badge"
                :class="item.policy?.enabled ? 'badge-success' : 'badge-warning'"
              >
                {{ item.policy?.enabled ? "已启用" : "未启用" }}
              </span>
            </div>
            <div class="mt-3 grid gap-2 md:grid-cols-3">
              <div class="sub-card">
                <p class="table-meta">下游协议</p>
                <p class="mt-2 font-medium text-slate-900">{{ item.key }}</p>
              </div>
              <div class="sub-card">
                <p class="table-meta">供应商</p>
                <p class="mt-2 font-medium text-slate-900">{{ item.policy?.supplier_name || "未分配" }}</p>
              </div>
              <div class="sub-card">
                <p class="table-meta">目标模型</p>
                <p class="mt-2 font-medium text-slate-900">{{ item.policy?.target_model || "未设置" }}</p>
              </div>
            </div>
            <div class="mt-3">
              <button
                class="btn btn-secondary"
                @click="item.policy ? openPolicyEdit(item.policy) : openPolicyCreate(item.key)"
              >
                {{ item.policy ? "编辑该路由" : "配置该路由" }}
              </button>
            </div>
          </article>
        </div>
      </PanelBlock>

      <PanelBlock title="默认路由策略">
        <div v-if="!store.policies.length" class="empty-state">
          当前尚未配置路由策略。
        </div>
        <div v-else class="table-shell">
          <div class="table-scroll">
            <table class="admin-table">
              <thead>
                <tr>
                  <th>下游协议</th>
                  <th>供应商</th>
                  <th>上游 / 模型</th>
                  <th>状态</th>
                  <th>操作</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="policy in store.policies" :key="policy.id">
                  <td>
                    <p class="font-medium text-slate-900">{{ policy.downstream_protocol }}</p>
                  </td>
                  <td>
                    <p class="text-sm text-slate-700">{{ policy.supplier_name || "未分配" }}</p>
                    <p class="mt-1 table-meta">{{ policy.supplier_id || "-" }}</p>
                  </td>
                  <td>
                    <p class="text-sm text-slate-700">{{ policy.upstream_protocol || "-" }}</p>
                    <div class="mt-2">
                      <code class="mono-chip">{{ policy.target_model || "无模型" }}</code>
                    </div>
                  </td>
                  <td>
                    <span class="badge" :class="policy.enabled ? 'badge-success' : 'badge-danger'">
                      {{ policy.enabled ? "启用" : "停用" }}
                    </span>
                  </td>
                  <td>
                    <button class="btn btn-secondary" @click="openPolicyEdit(policy)">
                      编辑策略
                    </button>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
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

        <FieldLabel label="描述">
          <textarea v-model="store.form.description" class="field-input min-h-24" placeholder="填写该供应商配置的用途说明" />
        </FieldLabel>

        <div class="grid gap-3 md:grid-cols-2">
          <FieldLabel label="模型列表">
            <input v-model="store.form.models" class="field-input" placeholder="gpt-4.1, gpt-4.1-mini" />
          </FieldLabel>
          <FieldLabel label="标签">
            <input v-model="store.form.tags" class="field-input" placeholder="official, primary" />
          </FieldLabel>
        </div>

        <label class="field-toggle">
          <input v-model="store.form.enabled" type="checkbox" class="field-checkbox" />
          启用该供应商配置
        </label>
      </form>
      <template #footer>
        <div class="flex justify-end gap-2">
          <button type="button" class="btn btn-secondary" @click="closeSupplierModal">取消</button>
          <button form="supplier-form" class="btn btn-primary" :disabled="store.saving">
            {{ store.saving ? "保存中..." : store.form.id ? "更新供应商" : "创建供应商" }}
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

        <FieldLabel label="目标模型">
          <input v-model="store.policyForm.target_model" class="field-input" placeholder="例如：gpt-4.1-mini 或 claude-sonnet-4" />
        </FieldLabel>

        <label class="field-toggle">
          <input v-model="store.policyForm.enabled" type="checkbox" class="field-checkbox" />
          启用该路由策略
        </label>
      </form>
      <template #footer>
        <div class="flex justify-end gap-2">
          <button type="button" class="btn btn-secondary" @click="closePolicyModal">取消</button>
          <button form="policy-form" class="btn btn-primary" :disabled="store.saving">
            {{ store.saving ? "保存中..." : "保存路由策略" }}
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

const store = useSuppliersStore();
const supplierModalOpen = ref(false);
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

function healthTone(record) {
  if (!record) {
    return { badge: "badge-neutral" };
  }
  if (record.status === "reachable") {
    return { badge: "badge-success" };
  }
  if (record.status === "warning") {
    return { badge: "badge-warning" };
  }
  return { badge: "badge-danger" };
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

async function submitSupplier() {
  await store.save();
  if (!store.error) {
    supplierModalOpen.value = false;
  }
}

function openPolicyCreate(protocol = "anthropic") {
  store.policyForm = {
    id: "",
    downstream_protocol: protocol,
    supplier_id: "",
    target_model: "",
    enabled: true,
  };
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

async function confirmDelete() {
  if (!confirmState.id) {
    return;
  }
  await store.remove(confirmState.id);
  confirmState.open = false;
  confirmState.id = "";
  confirmState.message = "";
}

onMounted(() => {
  store.load();
});
</script>
