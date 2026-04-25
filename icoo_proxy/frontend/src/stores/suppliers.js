import { defineStore } from "pinia";
import { DeleteSupplier, ListRoutePolicies, ListSuppliers, SaveRoutePolicy, SaveSupplier } from "../../wailsjs/go/main/App";

const emptyForm = () => ({
  id: "",
  name: "",
  protocol: "openai-responses",
  base_url: "",
  api_key: "",
  enabled: true,
  description: "",
  models: "",
  tags: "",
});

export const useSuppliersStore = defineStore("suppliers", {
  state: () => ({
    loading: false,
    saving: false,
    deleting: "",
    error: "",
    items: [],
    policies: [],
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
  },
  actions: {
    async load() {
      this.loading = true;
      this.error = "";
      try {
        const [items, policies] = await Promise.all([ListSuppliers(), ListRoutePolicies()]);
        this.items = items;
        this.policies = policies;
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
        enabled: Boolean(item.enabled),
        description: item.description || "",
        models: (item.models || []).join(", "),
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
    async save() {
      this.saving = true;
      this.error = "";
      try {
        this.items = await SaveSupplier({ ...this.form });
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
        if (this.form.id === id) {
          this.resetForm();
        }
      } catch (error) {
        this.error = error?.message || String(error);
      } finally {
        this.deleting = "";
      }
    },
  },
});
