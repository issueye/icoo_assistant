import { defineStore } from "pinia";
import { DeleteSupplier, ListSuppliers, SaveSupplier } from "../../wailsjs/go/main/App";

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
    form: emptyForm(),
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
        this.items = await ListSuppliers();
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
