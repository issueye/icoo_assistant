<template>
  <div class="min-h-screen text-slate-100">
    <div class="mx-auto flex min-h-screen max-w-[1600px] gap-6 px-4 py-5 lg:px-6">
      <aside class="hidden w-72 shrink-0 rounded-[28px] border border-white/10 bg-white/5 p-5 shadow-panel backdrop-blur lg:flex lg:flex-col">
        <div class="mb-8">
          <p class="text-xs uppercase tracking-[0.28em] text-signal-amber">icoo_proxy</p>
          <h1 class="mt-3 text-3xl font-bold tracking-[-0.04em]">Gateway Console</h1>
          <p class="mt-3 text-sm leading-6 text-slate-300/80">
            Desktop control plane for the local AI gateway, protocol translation, and supplier registry.
          </p>
        </div>

        <nav class="space-y-2">
          <RouterLink
            v-for="item in navItems"
            :key="item.to"
            :to="item.to"
            class="block rounded-2xl border px-4 py-3 transition"
            :class="route.path === item.to ? 'border-signal-mint/40 bg-signal-mint/10 text-white' : 'border-white/5 bg-black/10 text-slate-300 hover:border-white/15 hover:bg-white/5'"
          >
            <p class="text-sm font-semibold">{{ item.label }}</p>
            <p class="mt-1 text-xs text-slate-400">{{ item.description }}</p>
          </RouterLink>
        </nav>

        <div class="mt-auto rounded-3xl border border-white/10 bg-ink-900/80 p-4">
          <p class="text-xs uppercase tracking-[0.24em] text-signal-sky">Current Scope</p>
          <ul class="mt-3 space-y-2 text-sm text-slate-300/80">
            <li>Three protocol entrypoints</li>
            <li>Cross-protocol non-streaming translation</li>
            <li>Supplier registry with local persistence</li>
          </ul>
        </div>
      </aside>

      <main class="min-w-0 flex-1">
        <RouterView />
      </main>
    </div>
  </div>
</template>

<script setup>
import { computed } from "vue";
import { RouterLink, RouterView, useRoute } from "vue-router";

const route = useRoute();

const navItems = computed(() => [
  {
    to: "/",
    label: "Gateway Overview",
    description: "Runtime status, route catalog, and recent traffic.",
  },
  {
    to: "/suppliers",
    label: "Supplier Management",
    description: "Manage upstream vendors, endpoints, models, and tags.",
  },
]);
</script>
