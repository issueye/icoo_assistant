import { createRouter, createWebHashHistory } from "vue-router";

import OverviewView from "../views/OverviewView.vue";
import SuppliersView from "../views/SuppliersView.vue";
import TrafficView from "../views/TrafficView.vue";
import UedSpecView from "../views/UedSpecView.vue";

export default createRouter({
  history: createWebHashHistory(),
  routes: [
    {
      path: "/",
      name: "overview",
      component: OverviewView,
    },
    {
      path: "/suppliers",
      name: "suppliers",
      component: SuppliersView,
    },
    {
      path: "/traffic",
      name: "traffic",
      component: TrafficView,
    },
    {
      path: "/ued",
      name: "ued",
      component: UedSpecView,
    },
  ],
});
