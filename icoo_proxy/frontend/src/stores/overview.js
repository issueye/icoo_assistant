import { defineStore } from "pinia";
import { GetOverview, ReloadProxy } from "../../wailsjs/go/main/App";

export const useOverviewStore = defineStore("overview", {
  state: () => ({
    loading: false,
    refreshing: false,
    error: "",
    data: null,
  }),
  getters: {
    checks(state) {
      return state.data?.checks || {};
    },
    routes(state) {
      return state.data?.supported_paths || [];
    },
    requests(state) {
      return state.data?.recent_requests || [];
    },
  },
  actions: {
    async load() {
      this.loading = true;
      this.error = "";
      try {
        this.data = await GetOverview();
      } catch (error) {
        this.error = error?.message || String(error);
      } finally {
        this.loading = false;
      }
    },
    async reloadProxy() {
      this.refreshing = true;
      this.error = "";
      try {
        this.data = await ReloadProxy();
      } catch (error) {
        this.error = error?.message || String(error);
      } finally {
        this.refreshing = false;
      }
    },
  },
});
