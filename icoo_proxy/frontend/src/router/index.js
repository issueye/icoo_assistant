import { createRouter, createWebHashHistory } from "vue-router";

import OverviewView from "../views/OverviewView.vue";
import SuppliersView from "../views/SuppliersView.vue";

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
  ],
});
