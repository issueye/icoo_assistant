import { defineStore } from "pinia";
import { GetProjectSettings, SaveProjectSettings } from "../lib/wailsApp";

const emptyForm = () => ({
  proxy_host: "127.0.0.1",
  proxy_port: 18181,
  proxy_read_timeout_seconds: 15,
  proxy_write_timeout_seconds: 300,
  proxy_shutdown_timeout_seconds: 10,
  proxy_chain_log_path: ".data/icoo_proxy-chain.log",
  proxy_chain_log_bodies: true,
  proxy_chain_log_max_body_bytes: 0,
});

export const useSettingsStore = defineStore("settings", {
  state: () => ({
    loading: false,
    saving: false,
    error: "",
    success: "",
    form: emptyForm(),
  }),
  actions: {
    async load() {
      this.loading = true;
      this.error = "";
      this.success = "";
      try {
        const settings = await GetProjectSettings();
        this.form = { ...emptyForm(), ...settings };
      } catch (error) {
        this.error = error?.message || String(error);
      } finally {
        this.loading = false;
      }
    },
    async save() {
      this.saving = true;
      this.error = "";
      this.success = "";
      try {
        const payload = {
          ...this.form,
          proxy_port: Number(this.form.proxy_port),
          proxy_read_timeout_seconds: Number(this.form.proxy_read_timeout_seconds),
          proxy_write_timeout_seconds: Number(this.form.proxy_write_timeout_seconds),
          proxy_shutdown_timeout_seconds: Number(this.form.proxy_shutdown_timeout_seconds),
          proxy_chain_log_max_body_bytes: Number(this.form.proxy_chain_log_max_body_bytes),
        };
        const settings = await SaveProjectSettings(payload);
        this.form = { ...emptyForm(), ...settings };
        this.success = "设置已保存并自动重载代理。";
      } catch (error) {
        this.error = error?.message || String(error);
      } finally {
        this.saving = false;
      }
    },
    reset() {
      this.form = emptyForm();
      this.error = "";
      this.success = "";
    },
  },
});
