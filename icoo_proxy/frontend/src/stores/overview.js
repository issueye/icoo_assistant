import { defineStore } from "pinia";
import { GetOverview, ListSupplierHealth, ListSuppliers, ReloadProxy } from "../lib/wailsApp";

export const useOverviewStore = defineStore("overview", {
  state: () => ({
    loading: false,
    refreshing: false,
    error: "",
    data: null,
    suppliers: [],
    health: [],
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
    supplierCount(state) {
      return state.suppliers.length;
    },
    enabledSupplierCount(state) {
      return state.suppliers.filter((item) => item.enabled).length;
    },
    checkedSupplierCount(state) {
      return state.health.length;
    },
    reachableSupplierCount(state) {
      return state.health.filter((item) => item.status === "reachable").length;
    },
    warningSupplierCount(state) {
      return state.health.filter((item) => item.status !== "reachable").length;
    },
    activePolicyCount(state) {
      return (state.data?.route_policies || []).filter((item) => item.enabled).length;
    },
    inactivePolicyCount(state) {
      return (state.data?.route_policies || []).filter((item) => !item.enabled).length;
    },
    unhealthySuppliers(state) {
      return state.health.filter((item) => item.status !== "reachable");
    },
  },
  actions: {
    async load() {
      this.loading = true;
      this.error = "";
      try {
        const [overview, suppliers, health] = await Promise.all([GetOverview(), ListSuppliers(), ListSupplierHealth()]);
        this.data = overview;
        this.suppliers = suppliers;
        this.health = health;
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
        const [overview, suppliers, health] = await Promise.all([ReloadProxy(), ListSuppliers(), ListSupplierHealth()]);
        this.data = overview;
        this.suppliers = suppliers;
        this.health = health;
      } catch (error) {
        this.error = error?.message || String(error);
      } finally {
        this.refreshing = false;
      }
    },
  },
});
