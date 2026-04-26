import { defineStore } from "pinia";
import {
  CheckSupplier,
  DeleteSupplier,
  ListRoutePolicies,
  ListSupplierHealth,
  ListSuppliers,
  SaveRoutePolicy,
  SaveSupplier,
} from "../lib/wailsApp";

const emptyForm = () => ({
  id: "",
  name: "",
  protocol: "openai-responses",
  base_url: "",
  api_key: "",
  only_stream: false,
  user_agent: "",
  enabled: true,
  description: "",
  models: [""],
  tags: "",
});

export const useSuppliersStore = defineStore("suppliers", {
  state: () => ({
    loading: false,
    saving: false,
    deleting: "",
    checking: "",
    error: "",
    items: [],
    policies: [],
    health: [],
    form: emptyForm(),
    policyForm: {
      id: "",
      downstream_protocol: "anthropic",
      supplier_id: "",
      target_model: "",
      enabled: true,
    },
  }),
  getters: {
    enabledCount(state) {
      return state.items.filter((item) => item.enabled).length;
    },
    checkedCount(state) {
      return state.health.length;
    },
    routeDefinitions() {
      return [
        {
          key: "anthropic",
          label: "Anthropic 路由",
          description: "用于兼容 /v1/messages 与 /anthropic/v1/messages 请求。",
        },
        {
          key: "openai-chat",
          label: "Chat 路由",
          description: "用于兼容 /v1/chat/completions 与 /openai/v1/chat/completions 请求。",
        },
        {
          key: "openai-responses",
          label: "Responses 路由",
          description: "用于兼容 /v1/responses 与 /openai/v1/responses 请求。",
        },
      ];
    },
    policyOptions() {
      return this.routeDefinitions.map((item) => ({
        label: item.label,
        value: item.key,
      }));
    },
    policiesByProtocol() {
      const lookup = {};
      this.policies.forEach((item) => {
        lookup[item.downstream_protocol] = item;
      });
      return this.routeDefinitions.map((definition) => ({
        ...definition,
        policy: lookup[definition.key] || null,
      }));
    },
  },
  actions: {
    async load() {
      this.loading = true;
      this.error = "";
      try {
        const [items, policies, health] = await Promise.all([ListSuppliers(), ListRoutePolicies(), ListSupplierHealth()]);
        this.items = items;
        this.policies = policies;
        this.health = health;
      } catch (error) {
        this.error = error?.message || String(error);
      } finally {
        this.loading = false;
      }
    },
    select(item) {
      this.form = {
        id: item.id,
        name: item.name,
        protocol: item.protocol,
        base_url: item.base_url,
        api_key: "",
        only_stream: Boolean(item.only_stream),
        user_agent: item.user_agent || "",
        enabled: Boolean(item.enabled),
        description: item.description || "",
        models: item.models?.length ? [...item.models] : [""],
        tags: (item.tags || []).join(", "),
      };
    },
    resetForm() {
      this.form = emptyForm();
    },
    selectPolicy(item) {
      this.policyForm = {
        id: item.id,
        downstream_protocol: item.downstream_protocol,
        supplier_id: item.supplier_id,
        target_model: item.target_model || "",
        enabled: Boolean(item.enabled),
      };
    },
    resetPolicyForm() {
      this.policyForm = {
        id: "",
        downstream_protocol: "anthropic",
        supplier_id: "",
        target_model: "",
        enabled: true,
      };
    },
    healthFor(id) {
      return this.health.find((item) => item.supplier_id === id);
    },
    async save() {
      this.saving = true;
      this.error = "";
      try {
        this.items = await SaveSupplier({
          ...this.form,
          models: this.form.models.map((item) => String(item).trim()).filter(Boolean).join(", "),
        });
        this.resetForm();
      } catch (error) {
        this.error = error?.message || String(error);
      } finally {
        this.saving = false;
      }
    },
    async savePolicy() {
      this.saving = true;
      this.error = "";
      try {
        this.policies = await SaveRoutePolicy({ ...this.policyForm });
        this.resetPolicyForm();
      } catch (error) {
        this.error = error?.message || String(error);
      } finally {
        this.saving = false;
      }
    },
    async remove(id) {
      this.deleting = id;
      this.error = "";
      try {
        this.items = await DeleteSupplier(id);
        this.health = this.health.filter((item) => item.supplier_id !== id);
        if (this.form.id === id) {
          this.resetForm();
        }
      } catch (error) {
        this.error = error?.message || String(error);
      } finally {
        this.deleting = "";
      }
    },
    async check(id) {
      this.checking = id;
      this.error = "";
      try {
        this.health = await CheckSupplier(id);
      } catch (error) {
        this.error = error?.message || String(error);
      } finally {
        this.checking = "";
      }
    },
  },
});
