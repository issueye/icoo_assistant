<template>
  <section class="space-y-6">
    <div class="rounded-[30px] border border-white/10 bg-white/5 p-6 shadow-panel backdrop-blur">
      <div class="grid gap-6 xl:grid-cols-[1.2fr_0.8fr]">
        <div>
          <p class="text-xs uppercase tracking-[0.26em] text-signal-amber">Supplier Management</p>
          <h2 class="mt-3 text-4xl font-bold tracking-[-0.05em]">Manage upstream vendors and routing candidates.</h2>
          <p class="mt-4 max-w-3xl text-sm leading-7 text-slate-300/80">
            Supplier profiles are stored locally and can be used as the management base for future dynamic route
            selection, policy assignment, and provider-specific health checks.
          </p>
        </div>

        <div class="grid gap-3 sm:grid-cols-2 xl:grid-cols-1">
          <StatCard label="Total Suppliers" :value="String(store.items.length)" />
          <StatCard label="Enabled Profiles" :value="String(store.enabledCount)" />
          <StatCard
            label="Current Form"
            :value="store.form.id ? 'Editing existing supplier' : 'Creating new supplier'"
          />
        </div>
      </div>
    </div>

    <div v-if="store.error" class="rounded-3xl border border-signal-coral/25 bg-signal-coral/10 px-5 py-4 text-sm text-rose-100">
      {{ store.error }}
    </div>

    <div class="grid gap-6 xl:grid-cols-[1.1fr_0.9fr]">
      <PanelBlock title="Supplier Registry" eyebrow="Catalog">
        <div v-if="store.loading" class="rounded-3xl border border-white/10 bg-black/20 px-4 py-8 text-center text-sm text-slate-400">
          Loading suppliers...
        </div>
        <div v-else class="space-y-3">
          <article
            v-for="item in store.items"
            :key="item.id"
            class="rounded-3xl border border-white/10 bg-ink-900/70 p-4"
          >
            <div class="flex flex-wrap items-start justify-between gap-3">
              <div>
                <div class="flex items-center gap-2">
                  <p class="text-base font-semibold">{{ item.name }}</p>
                  <span
                    class="rounded-full px-2.5 py-1 text-[11px] font-semibold uppercase tracking-[0.18em]"
                    :class="item.enabled ? 'bg-signal-mint/15 text-signal-mint' : 'bg-signal-coral/15 text-signal-coral'"
                  >
                    {{ item.enabled ? "enabled" : "disabled" }}
                  </span>
                </div>
                <p class="mt-1 text-sm text-slate-400">{{ item.protocol }} | {{ item.base_url }}</p>
              </div>
              <div class="flex gap-2">
                <button
                  class="rounded-full border border-white/10 bg-white/5 px-4 py-2 text-xs font-semibold hover:border-white/20 hover:bg-white/10"
                  @click="store.select(item)"
                >
                  Edit
                </button>
                <button
                  class="rounded-full border border-signal-coral/20 bg-signal-coral/10 px-4 py-2 text-xs font-semibold text-rose-100 hover:bg-signal-coral/20 disabled:cursor-progress disabled:opacity-70"
                  :disabled="store.deleting === item.id"
                  @click="store.remove(item.id)"
                >
                  {{ store.deleting === item.id ? "Deleting..." : "Delete" }}
                </button>
              </div>
            </div>

            <p class="mt-3 text-sm leading-6 text-slate-300/80">{{ item.description || "No description yet." }}</p>

            <div class="mt-4 flex flex-wrap gap-2">
              <code class="rounded-full bg-black/20 px-3 py-1 font-mono text-xs">{{ item.api_key_masked || "No API key stored" }}</code>
              <span
                v-for="model in item.models || []"
                :key="model"
                class="rounded-full bg-signal-sky/10 px-3 py-1 text-xs text-signal-sky"
              >
                {{ model }}
              </span>
              <span
                v-for="tag in item.tags || []"
                :key="tag"
                class="rounded-full bg-white/5 px-3 py-1 text-xs text-slate-300"
              >
                #{{ tag }}
              </span>
            </div>
          </article>
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

          <label class="flex items-center gap-3 rounded-2xl border border-white/10 bg-black/10 px-4 py-3 text-sm text-slate-200">
            <input v-model="store.form.enabled" type="checkbox" class="h-4 w-4 rounded border-white/20 bg-black/20 text-signal-mint" />
            Enable this supplier profile
          </label>

          <div class="flex flex-wrap gap-3">
            <button
              class="rounded-full bg-signal-mint px-5 py-3 text-sm font-semibold text-ink-950 transition hover:-translate-y-0.5 disabled:cursor-progress disabled:opacity-70"
              :disabled="store.saving"
            >
              {{ store.saving ? "Saving..." : store.form.id ? "Update Supplier" : "Create Supplier" }}
            </button>
            <button
              type="button"
              class="rounded-full border border-white/10 bg-white/5 px-5 py-3 text-sm font-semibold hover:border-white/20 hover:bg-white/10"
              @click="store.resetForm"
            >
              Reset Form
            </button>
          </div>
        </form>
      </PanelBlock>
    </div>

    <div class="grid gap-6 xl:grid-cols-[1fr_1fr]">
      <PanelBlock title="Default Route Policies" eyebrow="Gateway Routing">
        <div class="space-y-3">
          <article
            v-for="policy in store.policies"
            :key="policy.id"
            class="rounded-3xl border border-white/10 bg-ink-900/70 p-4"
          >
            <div class="flex flex-wrap items-start justify-between gap-3">
              <div>
                <p class="text-sm font-semibold">{{ policy.downstream_protocol }}</p>
                <p class="mt-1 text-xs text-slate-400">{{ policy.supplier_name || "Unassigned" }} | {{ policy.upstream_protocol || "-" }}</p>
              </div>
              <button
                class="rounded-full border border-white/10 bg-white/5 px-4 py-2 text-xs font-semibold hover:border-white/20 hover:bg-white/10"
                @click="store.selectPolicy(policy)"
              >
                Edit Policy
              </button>
            </div>
            <div class="mt-3 flex flex-wrap gap-2">
              <code class="rounded-full bg-black/20 px-3 py-1 font-mono text-xs">{{ policy.target_model || "No model" }}</code>
              <span
                class="rounded-full px-3 py-1 text-[11px] font-semibold uppercase tracking-[0.18em]"
                :class="policy.enabled ? 'bg-signal-mint/15 text-signal-mint' : 'bg-signal-coral/15 text-signal-coral'"
              >
                {{ policy.enabled ? "enabled" : "disabled" }}
              </span>
            </div>
          </article>
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

          <label class="flex items-center gap-3 rounded-2xl border border-white/10 bg-black/10 px-4 py-3 text-sm text-slate-200">
            <input v-model="store.policyForm.enabled" type="checkbox" class="h-4 w-4 rounded border-white/20 bg-black/20 text-signal-mint" />
            Enable this route policy
          </label>

          <div class="flex flex-wrap gap-3">
            <button
              class="rounded-full bg-signal-amber px-5 py-3 text-sm font-semibold text-ink-950 transition hover:-translate-y-0.5 disabled:cursor-progress disabled:opacity-70"
              :disabled="store.saving"
            >
              {{ store.saving ? "Saving..." : "Save Route Policy" }}
            </button>
            <button
              type="button"
              class="rounded-full border border-white/10 bg-white/5 px-5 py-3 text-sm font-semibold hover:border-white/20 hover:bg-white/10"
              @click="store.resetPolicyForm"
            >
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

onMounted(() => {
  store.load();
});
</script>
