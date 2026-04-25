import { defineStore } from "pinia";
import { GetOverview } from "../lib/wailsApp";

export const useTrafficStore = defineStore("traffic", {
  state: () => ({
    loading: false,
    refreshing: false,
    error: "",
    requests: [],
    filter: "all",
    autoRefresh: false,
    lastUpdatedAt: "",
  }),
  getters: {
    filteredRequests(state) {
      if (state.filter === "all") {
        return state.requests;
      }
      return state.requests.filter((item) => item.downstream === state.filter || item.upstream === state.filter);
    },
    successCount(state) {
      return state.requests.filter((item) => item.status_code > 0 && item.status_code < 400).length;
    },
    errorCount(state) {
      return state.requests.filter((item) => item.status_code >= 400).length;
    },
    averageLatency(state) {
      if (state.requests.length === 0) {
        return 0;
      }
      const total = state.requests.reduce((sum, item) => sum + (item.duration_ms || 0), 0);
      return Math.round(total / state.requests.length);
    },
    protocolOptions(state) {
      const values = new Set(["all"]);
      state.requests.forEach((item) => {
        if (item.downstream) {
          values.add(item.downstream);
        }
        if (item.upstream) {
          values.add(item.upstream);
        }
      });
      return Array.from(values);
    },
  },
  actions: {
    async load() {
      this.loading = true;
      this.error = "";
      try {
        const overview = await GetOverview();
        this.requests = overview?.recent_requests || [];
        this.lastUpdatedAt = new Date().toISOString();
      } catch (error) {
        this.error = error?.message || String(error);
      } finally {
        this.loading = false;
      }
    },
    async refresh() {
      this.refreshing = true;
      this.error = "";
      try {
        const overview = await GetOverview();
        this.requests = overview?.recent_requests || [];
        this.lastUpdatedAt = new Date().toISOString();
      } catch (error) {
        this.error = error?.message || String(error);
      } finally {
        this.refreshing = false;
      }
    },
    setFilter(value) {
      this.filter = value;
    },
    toggleAutoRefresh() {
      this.autoRefresh = !this.autoRefresh;
    },
  },
});
