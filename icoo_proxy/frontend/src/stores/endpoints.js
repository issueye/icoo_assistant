import { defineStore } from "pinia";
import { DeleteEndpoint, ListEndpoints, ReloadProxy, SaveEndpoint } from "../lib/wailsApp";

const emptyForm = () => ({
  id: "",
  path: "",
  protocol: "openai-chat",
  description: "",
  enabled: true,
});

export const useEndpointsStore = defineStore("endpoints", {
  state: () => ({
    loading: false,
    saving: false,
    deleting: "",
    reloading: false,
    error: "",
    items: [],
    form: emptyForm(),
  }),
  getters: {
    enabledCount(state) {
      return state.items.filter((item) => item.enabled).length;
    },
    customCount(state) {
      return state.items.filter((item) => !item.built_in).length;
    },
    protocolOptions() {
      return [
        { label: "anthropic", value: "anthropic" },
        { label: "openai-chat", value: "openai-chat" },
        { label: "openai-responses", value: "openai-responses" },
      ];
    },
  },
  actions: {
    async load() {
      this.loading = true;
      this.error = "";
      try {
        this.items = await ListEndpoints();
      } catch (error) {
        this.error = error?.message || String(error);
      } finally {
        this.loading = false;
      }
    },
    select(item) {
      this.form = {
        id: item.id,
        path: item.path,
        protocol: item.protocol,
        description: item.description || "",
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
        this.items = await SaveEndpoint({ ...this.form });
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
        this.items = await DeleteEndpoint(id);
      } catch (error) {
        this.error = error?.message || String(error);
      } finally {
        this.deleting = "";
      }
    },
    async reloadProxy() {
      this.reloading = true;
      this.error = "";
      try {
        await ReloadProxy();
      } catch (error) {
        this.error = error?.message || String(error);
      } finally {
        this.reloading = false;
      }
    },
  },
});
