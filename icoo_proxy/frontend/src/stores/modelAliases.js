import { defineStore } from "pinia";
import {
  DeleteModelAlias,
  ListModelAliases,
  ListSuppliers,
  SaveModelAlias,
} from "../lib/wailsApp";

const emptyForm = () => ({
  id: "",
  name: "",
  supplier_id: "",
  model: "",
  enabled: true,
});

export const useModelAliasesStore = defineStore("modelAliases", {
  state: () => ({
    loading: false,
    saving: false,
    deleting: "",
    error: "",
    items: [],
    suppliers: [],
    form: emptyForm(),
  }),
  getters: {
    enabledCount(state) {
      return state.items.filter((item) => item.enabled).length;
    },
    supplierCount(state) {
      const seen = new Set();
      state.items.forEach((item) => {
        if (item.supplier_id) {
          seen.add(item.supplier_id);
        }
      });
      return seen.size;
    },
    supplierOptions(state) {
      return state.suppliers.map((supplier) => ({
        label: `${supplier.name} (${supplier.protocol})`,
        value: supplier.id,
      }));
    },
    selectedSupplier(state) {
      return state.suppliers.find(
        (supplier) => supplier.id === state.form.supplier_id,
      ) || null;
    },
    modelOptions(state) {
      const supplier = state.suppliers.find(
        (s) => s.id === state.form.supplier_id,
      );
      if (!supplier || !supplier.models?.length) {
        return [];
      }
      return supplier.models.map((model) => ({
        label:
          model === supplier.default_model ? `${model} (默认)` : model,
        value: model,
      }));
    },
  },
  actions: {
    async load() {
      this.loading = true;
      this.error = "";
      try {
        const [items, suppliers] = await Promise.all([
          ListModelAliases(),
          ListSuppliers(),
        ]);
        this.items = items;
        this.suppliers = suppliers;
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
        supplier_id: item.supplier_id || "",
        model: item.model,
        enabled: Boolean(item.enabled),
      };
    },
    resetForm() {
      this.form = emptyForm();
    },
    async save() {
      this.saving = true;
      this.error = "";
      try {
        this.items = await SaveModelAlias({
          id: this.form.id,
          name: String(this.form.name || "").trim(),
          supplier_id: String(this.form.supplier_id || "").trim(),
          model: String(this.form.model || "").trim(),
          enabled: this.form.enabled,
        });
        this.resetForm();
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
        this.items = await DeleteModelAlias(id);
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
