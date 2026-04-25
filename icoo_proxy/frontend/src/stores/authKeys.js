import { defineStore } from "pinia";
import { DeleteAuthKey, GetAuthKeySecret, ListAuthKeys, ReloadProxy, SaveAuthKey } from "../lib/wailsApp";

const emptyForm = () => ({
  id: "",
  name: "",
  secret: "",
  enabled: true,
  description: "",
});

function randomSecret() {
  const bytes = new Uint8Array(24);
  window.crypto?.getRandomValues?.(bytes);
  const hex = Array.from(bytes, (item) => item.toString(16).padStart(2, "0")).join("");
  return `icoo_${hex || Date.now()}`;
}

export const useAuthKeysStore = defineStore("authKeys", {
  state: () => ({
    loading: false,
    saving: false,
    deleting: "",
    reloading: false,
    copying: "",
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
        this.items = await ListAuthKeys();
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
        secret: "",
        enabled: Boolean(item.enabled),
        description: item.description || "",
      };
    },
    resetForm() {
      this.form = emptyForm();
    },
    generateSecret() {
      this.form.secret = randomSecret();
    },
    async save() {
      this.saving = true;
      this.error = "";
      try {
        this.items = await SaveAuthKey({ ...this.form });
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
        this.items = await DeleteAuthKey(id);
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
    async copySecret(id) {
      this.copying = id;
      this.error = "";
      try {
        const secret = await GetAuthKeySecret(id);
        if (secret) {
          await navigator.clipboard.writeText(secret);
        }
        return secret;
      } catch (error) {
        this.error = error?.message || String(error);
        return "";
      } finally {
        this.copying = "";
      }
    },
  },
});
