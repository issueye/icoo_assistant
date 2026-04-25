<template>
  <section class="page-section">
    <div class="page-header">
      <p class="page-eyebrow">Supplier Management</p>
      <h2 class="page-title">供应商与路由策略管理</h2>
      <p class="page-description">
        在这里集中维护上游供应商、目标模型、健康检查结果和默认路由策略，界面交互调整为传统后台管理台风格，便于录入、查询和维护。
      </p>
    </div>

    <div v-if="store.error" class="notice-error">
      {{ store.error }}
    </div>

    <div class="section-grid xl:grid-cols-4">
      <StatCard label="Total Suppliers" :value="String(store.items.length)" />
      <StatCard label="Enabled Profiles" :value="String(store.enabledCount)" />
      <StatCard label="Health Checked" :value="String(store.checkedCount)" />
      <StatCard label="Current Form" :value="store.form.id ? 'Editing existing supplier' : 'Creating new supplier'" />
    </div>

    <div class="section-grid xl:grid-cols-[1.15fr_0.85fr]">
      <PanelBlock title="Supplier Registry" eyebrow="Catalog">
        <div v-if="store.loading" class="empty-state">
          Loading suppliers...
        </div>
        <div v-else-if="!store.items.length" class="empty-state">
          No suppliers configured yet.
        </div>
        <div v-else class="table-shell">
          <div class="table-scroll">
            <table class="admin-table">
              <thead>
                <tr>
                  <th>Supplier</th>
                  <th>Protocol / Endpoint</th>
                  <th>Models / Tags</th>
                  <th>Health</th>
                  <th>Actions</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="item in store.items" :key="item.id">
                  <td>
                    <div class="flex items-center gap-2">
                      <p class="font-medium text-slate-900">{{ item.name }}</p>
                      <span class="badge" :class="item.enabled ? 'badge-success' : 'badge-danger'">
                        {{ item.enabled ? "enabled" : "disabled" }}
                      </span>
                    </div>
                    <p class="mt-2 text-sm text-slate-600">{{ item.description || "No description yet." }}</p>
                    <p class="mt-2 table-meta">Updated: {{ formatCheckedAt(item.updated_at) }}</p>
                  </td>
                  <td>
                    <p class="font-medium text-slate-900">{{ item.protocol }}</p>
                    <p class="mt-2 break-all table-meta">{{ item.base_url }}</p>
                    <div class="mt-2">
                      <code class="mono-chip">{{ item.api_key_masked || "No API key stored" }}</code>
                    </div>
                  </td>
                  <td>
                    <div class="flex flex-wrap gap-2">
                      <span v-for="model in item.models || []" :key="model" class="tag-chip">
                        {{ model }}
                      </span>
                      <span v-if="!(item.models || []).length" class="table-meta">No models</span>
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
                        HTTP {{ store.healthFor(item.id).status_code || "no-status" }}
                      </p>
                      <p class="mt-2 text-sm text-slate-600">{{ store.healthFor(item.id).message }}</p>
                      <p class="mt-2 table-meta">{{ formatCheckedAt(store.healthFor(item.id).checked_at) }}</p>
                    </template>
                    <span v-else class="table-meta">Not checked yet</span>
                  </td>
                  <td>
                    <div class="table-actions">
                      <button class="btn btn-info" :disabled="store.checking === item.id" @click="store.check(item.id)">
                        {{ store.checking === item.id ? "Checking..." : "Check" }}
                      </button>
                      <button class="btn btn-secondary" @click="store.select(item)">Edit</button>
                      <button class="btn btn-danger" :disabled="store.deleting === item.id" @click="store.remove(item.id)">
                        {{ store.deleting === item.id ? "Deleting..." : "Delete" }}
                      </button>
                    </div>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </PanelBlock>

      <PanelBlock title="Supplier Form" eyebrow="Editor">
        <form class="space-y-4" @submit.prevent="store.save">
          <div class="grid gap-4 md:grid-cols-2">
            <FieldLabel label="Name">
              <input v-model="store.form.name" class="field-input" placeholder="OpenAI Production" />
            </FieldLabel>
            <FieldLabel label="Protocol">
              <select v-model="store.form.protocol" class="field-input">
                <option value="anthropic">anthropic</option>
                <option value="openai-chat">openai-chat</option>
                <option value="openai-responses">openai-responses</option>
              </select>
            </FieldLabel>
          </div>

          <FieldLabel label="Base URL">
            <input v-model="store.form.base_url" class="field-input" placeholder="https://api.openai.com" />
          </FieldLabel>

          <FieldLabel label="API Key">
            <input v-model="store.form.api_key" class="field-input" placeholder="Leave blank to keep existing key on edit" />
          </FieldLabel>

          <FieldLabel label="Description">
            <textarea v-model="store.form.description" class="field-input min-h-24" placeholder="Describe what this supplier profile is used for." />
          </FieldLabel>

          <div class="grid gap-4 md:grid-cols-2">
            <FieldLabel label="Models">
              <input v-model="store.form.models" class="field-input" placeholder="gpt-4.1, gpt-4.1-mini" />
            </FieldLabel>
            <FieldLabel label="Tags">
              <input v-model="store.form.tags" class="field-input" placeholder="official, primary" />
            </FieldLabel>
          </div>

          <label class="field-toggle">
            <input v-model="store.form.enabled" type="checkbox" class="field-checkbox" />
            Enable this supplier profile
          </label>

          <div class="flex flex-wrap gap-3">
            <button class="btn btn-primary" :disabled="store.saving">
              {{ store.saving ? "Saving..." : store.form.id ? "Update Supplier" : "Create Supplier" }}
            </button>
            <button type="button" class="btn btn-secondary" @click="store.resetForm">
              Reset Form
            </button>
          </div>
        </form>
      </PanelBlock>
    </div>

    <div class="section-grid xl:grid-cols-2">
      <PanelBlock title="Default Route Policies" eyebrow="Gateway Routing">
        <div v-if="!store.policies.length" class="empty-state">
          No route policies configured yet.
        </div>
        <div v-else class="table-shell">
          <div class="table-scroll">
            <table class="admin-table">
              <thead>
                <tr>
                  <th>Downstream</th>
                  <th>Supplier</th>
                  <th>Upstream / Model</th>
                  <th>Status</th>
                  <th>Actions</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="policy in store.policies" :key="policy.id">
                  <td>
                    <p class="font-medium text-slate-900">{{ policy.downstream_protocol }}</p>
                  </td>
                  <td>
                    <p class="text-sm text-slate-700">{{ policy.supplier_name || "Unassigned" }}</p>
                    <p class="mt-1 table-meta">{{ policy.supplier_id || "-" }}</p>
                  </td>
                  <td>
                    <p class="text-sm text-slate-700">{{ policy.upstream_protocol || "-" }}</p>
                    <div class="mt-2">
                      <code class="mono-chip">{{ policy.target_model || "No model" }}</code>
                    </div>
                  </td>
                  <td>
                    <span class="badge" :class="policy.enabled ? 'badge-success' : 'badge-danger'">
                      {{ policy.enabled ? "enabled" : "disabled" }}
                    </span>
                  </td>
                  <td>
                    <button class="btn btn-secondary" @click="store.selectPolicy(policy)">
                      Edit Policy
                    </button>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </PanelBlock>

      <PanelBlock title="Route Policy Form" eyebrow="Default Binding">
        <form class="space-y-4" @submit.prevent="store.savePolicy">
          <div class="grid gap-4 md:grid-cols-2">
            <FieldLabel label="Downstream Protocol">
              <select v-model="store.policyForm.downstream_protocol" class="field-input">
                <option value="anthropic">anthropic</option>
                <option value="openai-chat">openai-chat</option>
                <option value="openai-responses">openai-responses</option>
              </select>
            </FieldLabel>
            <FieldLabel label="Supplier">
              <select v-model="store.policyForm.supplier_id" class="field-input">
                <option value="">Select supplier</option>
                <option v-for="supplier in store.items" :key="supplier.id" :value="supplier.id">
                  {{ supplier.name }} ({{ supplier.protocol }})
                </option>
              </select>
            </FieldLabel>
          </div>

          <FieldLabel label="Target Model">
            <input v-model="store.policyForm.target_model" class="field-input" placeholder="gpt-4.1-mini or claude-sonnet-4" />
          </FieldLabel>

          <label class="field-toggle">
            <input v-model="store.policyForm.enabled" type="checkbox" class="field-checkbox" />
            Enable this route policy
          </label>

          <div class="flex flex-wrap gap-3">
            <button class="btn btn-primary" :disabled="store.saving">
              {{ store.saving ? "Saving..." : "Save Route Policy" }}
            </button>
            <button type="button" class="btn btn-secondary" @click="store.resetPolicyForm">
              Reset Policy Form
            </button>
          </div>
        </form>
      </PanelBlock>
    </div>
  </section>
</template>

<script setup>
import { onMounted } from "vue";
import { useSuppliersStore } from "../stores/suppliers";

import FieldLabel from "../components/FieldLabel.vue";
import PanelBlock from "../components/PanelBlock.vue";
import StatCard from "../components/StatCard.vue";

const store = useSuppliersStore();

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
    return "Not checked yet";
  }
  return new Date(value).toLocaleString();
}

onMounted(() => {
  store.load();
});
</script>
