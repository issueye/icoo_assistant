<template>
  <section class="space-y-6">
    <div class="rounded-[30px] border border-white/10 bg-white/5 p-6 shadow-panel backdrop-blur">
      <div class="grid gap-6 xl:grid-cols-[1.3fr_0.9fr]">
        <div>
          <p class="text-xs uppercase tracking-[0.26em] text-signal-amber">Gateway Overview</p>
          <h2 class="mt-3 max-w-2xl text-4xl font-bold tracking-[-0.05em]">
            One local entrypoint for Anthropic, Chat, Responses, and managed suppliers.
          </h2>
          <p class="mt-4 max-w-3xl text-sm leading-7 text-slate-300/80">
            This dashboard keeps runtime health, route catalog, and recent traffic visible while the Go backend
            handles protocol translation, request forwarding, and now supplier registry management.
          </p>
          <div class="mt-6 flex flex-wrap items-center gap-3">
            <button
              class="rounded-full bg-signal-amber px-5 py-3 text-sm font-semibold text-ink-950 transition hover:-translate-y-0.5 disabled:cursor-progress disabled:opacity-70"
              :disabled="store.refreshing"
              @click="store.reloadProxy"
            >
              {{ store.refreshing ? "Reloading..." : "Reload Proxy" }}
            </button>
            <span
              class="rounded-full px-4 py-2 text-sm font-semibold"
              :class="store.data?.running ? 'bg-signal-mint/15 text-signal-mint' : 'bg-signal-coral/15 text-signal-coral'"
            >
              {{ store.data?.running ? "Running" : "Stopped" }}
            </span>
          </div>
        </div>

        <div class="grid gap-3 sm:grid-cols-2 xl:grid-cols-1">
          <StatCard label="Proxy URL" :value="store.data?.proxy_url || 'Not running'" />
          <StatCard label="Listen Address" :value="store.data?.listen_addr || '-'" />
          <StatCard label="Access Mode" :value="store.data?.auth_required ? 'API key required' : 'Trusted local mode'" />
          <StatCard label="Version" :value="store.data?.version || '-'" />
        </div>
      </div>
    </div>

    <div v-if="store.error" class="rounded-3xl border border-signal-coral/25 bg-signal-coral/10 px-5 py-4 text-sm text-rose-100">
      {{ store.error }}
    </div>

    <div v-if="store.loading" class="rounded-3xl border border-white/10 bg-black/20 px-5 py-10 text-center text-sm text-slate-300">
      Loading gateway overview...
    </div>

    <template v-else>
      <div class="grid gap-6 xl:grid-cols-2">
        <PanelBlock title="Provider Readiness" eyebrow="Runtime">
          <div class="grid gap-3 md:grid-cols-3">
            <article
              v-for="upstream in store.data?.upstreams || []"
              :key="upstream.protocol"
              class="rounded-3xl border border-white/10 bg-ink-900/70 p-4"
            >
              <p class="text-xs uppercase tracking-[0.22em] text-signal-sky">{{ upstream.protocol }}</p>
              <p class="mt-3 text-lg font-semibold">{{ upstream.configured ? "Configured" : "Missing key" }}</p>
              <p class="mt-2 break-all text-sm text-slate-400">{{ upstream.base_url || "-" }}</p>
            </article>
          </div>
          <div class="mt-4 flex flex-wrap gap-2">
            <span
              v-for="(value, key) in store.checks"
              :key="key"
              class="rounded-full px-3 py-2 text-xs font-semibold"
              :class="value ? 'bg-signal-mint/15 text-signal-mint' : 'bg-signal-amber/15 text-signal-amber'"
            >
              {{ key }}: {{ value }}
            </span>
          </div>
        </PanelBlock>

        <PanelBlock title="Supported Endpoints" eyebrow="Surface Area">
          <div class="flex flex-wrap gap-2">
            <code
              v-for="route in store.routes"
              :key="route"
              class="rounded-full bg-ink-900/90 px-3 py-2 font-mono text-xs text-slate-100"
            >
              {{ route }}
            </code>
          </div>
        </PanelBlock>
      </div>

      <div class="grid gap-6 xl:grid-cols-2">
        <PanelBlock title="Default Routes" eyebrow="Routing">
          <RouteList :items="store.data?.defaults || []" empty-text="No default routes configured yet." />
        </PanelBlock>

        <PanelBlock title="Model Aliases" eyebrow="Catalog">
          <RouteList :items="store.data?.aliases || []" empty-text="No aliases configured yet." />
        </PanelBlock>
      </div>

      <PanelBlock title="Recent Requests" eyebrow="Traffic">
        <div v-if="store.requests.length === 0" class="rounded-3xl border border-white/10 bg-black/20 px-4 py-8 text-center text-sm text-slate-400">
          No requests recorded yet.
        </div>
        <div v-else class="space-y-3">
          <article
            v-for="request in store.requests"
            :key="request.request_id"
            class="rounded-3xl border border-white/10 bg-ink-900/70 p-4"
          >
            <div class="flex flex-wrap items-start justify-between gap-3">
              <div>
                <p class="text-sm font-semibold">{{ request.downstream }} -> {{ request.upstream || "-" }}</p>
                <p class="mt-1 text-xs text-slate-400">{{ request.request_id }} | {{ request.created_at }}</p>
              </div>
              <div class="flex flex-wrap items-center gap-2">
                <code class="rounded-full bg-black/20 px-3 py-1 font-mono text-xs">{{ request.model || "-" }}</code>
                <span
                  class="rounded-full px-3 py-1 text-xs font-semibold"
                  :class="request.status_code >= 400 ? 'bg-signal-coral/15 text-signal-coral' : 'bg-signal-mint/15 text-signal-mint'"
                >
                  {{ request.status_code }}
                </span>
                <span class="rounded-full bg-white/5 px-3 py-1 text-xs text-slate-300">{{ request.duration_ms }} ms</span>
              </div>
            </div>
            <p v-if="request.error" class="mt-3 text-sm text-rose-200">{{ request.error }}</p>
          </article>
        </div>
      </PanelBlock>

      <PanelBlock title="Current Build Notes" eyebrow="Scope">
        <ul class="space-y-2 text-sm text-slate-300/80">
          <li v-for="note in store.data?.notes || []" :key="note">{{ note }}</li>
        </ul>
      </PanelBlock>
    </template>
  </section>
</template>

<script setup>
import { onMounted } from "vue";
import { useOverviewStore } from "../stores/overview";

import PanelBlock from "../components/PanelBlock.vue";
import RouteList from "../components/RouteList.vue";
import StatCard from "../components/StatCard.vue";

const store = useOverviewStore();

onMounted(() => {
  store.load();
});
</script>
