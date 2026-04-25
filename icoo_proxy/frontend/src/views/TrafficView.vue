<template>
  <section class="space-y-6">
    <div class="rounded-[30px] border border-white/10 bg-white/5 p-6 shadow-panel backdrop-blur">
      <div class="grid gap-6 xl:grid-cols-[1.25fr_0.75fr]">
        <div>
          <p class="text-xs uppercase tracking-[0.26em] text-signal-amber">Traffic Monitor</p>
          <h2 class="mt-3 text-4xl font-bold tracking-[-0.05em]">Review recent gateway traffic without leaving the desktop console.</h2>
          <p class="mt-4 max-w-3xl text-sm leading-7 text-slate-300/80">
            This view focuses on the latest proxied requests so we can quickly spot failures, protocol mismatches,
            and latency spikes while iterating on routing and supplier configuration.
          </p>
          <div class="mt-6 flex flex-wrap items-center gap-3">
            <button
              class="rounded-full bg-signal-amber px-5 py-3 text-sm font-semibold text-ink-950 transition hover:-translate-y-0.5 disabled:cursor-progress disabled:opacity-70"
              :disabled="store.refreshing"
              @click="store.refresh"
            >
              {{ store.refreshing ? "Refreshing..." : "Refresh Traffic" }}
            </button>
            <label class="flex items-center gap-3 rounded-full border border-white/10 bg-black/10 px-4 py-2 text-sm text-slate-200">
              <input
                :checked="store.autoRefresh"
                type="checkbox"
                class="h-4 w-4 rounded border-white/20 bg-black/20 text-signal-mint"
                @change="store.toggleAutoRefresh"
              />
              Auto refresh every 6s
            </label>
          </div>
        </div>

        <div class="grid gap-3 sm:grid-cols-2 xl:grid-cols-1">
          <StatCard label="Recent Requests" :value="String(store.requests.length)" />
          <StatCard label="Successful" :value="String(store.successCount)" />
          <StatCard label="Errors" :value="String(store.errorCount)" />
          <StatCard label="Avg Latency" :value="`${store.averageLatency} ms`" />
        </div>
      </div>
    </div>

    <div v-if="store.error" class="rounded-3xl border border-signal-coral/25 bg-signal-coral/10 px-5 py-4 text-sm text-rose-100">
      {{ store.error }}
    </div>

    <div class="grid gap-6 xl:grid-cols-[0.85fr_1.15fr]">
      <PanelBlock title="Filters" eyebrow="Focus">
        <div class="space-y-4">
          <label class="block">
            <span class="mb-2 block text-sm font-medium text-slate-200">Protocol</span>
            <select :value="store.filter" class="field-input" @change="store.setFilter($event.target.value)">
              <option v-for="option in store.protocolOptions" :key="option" :value="option">
                {{ option }}
              </option>
            </select>
          </label>

          <div class="rounded-3xl border border-white/10 bg-black/20 p-4 text-sm text-slate-300/80">
            <p class="font-semibold text-slate-100">Last updated</p>
            <p class="mt-2 text-slate-400">{{ formatDateTime(store.lastUpdatedAt) }}</p>
          </div>

          <div class="rounded-3xl border border-white/10 bg-black/20 p-4 text-sm text-slate-300/80">
            <p class="font-semibold text-slate-100">Current filter result</p>
            <p class="mt-2 text-slate-400">{{ store.filteredRequests.length }} request(s) visible</p>
          </div>
        </div>
      </PanelBlock>

      <PanelBlock title="Recent Request Timeline" eyebrow="Inspection">
        <div v-if="store.loading" class="rounded-3xl border border-white/10 bg-black/20 px-4 py-8 text-center text-sm text-slate-400">
          Loading traffic...
        </div>
        <div v-else-if="store.filteredRequests.length === 0" class="rounded-3xl border border-white/10 bg-black/20 px-4 py-8 text-center text-sm text-slate-400">
          No matching requests yet.
        </div>
        <div v-else class="space-y-3">
          <article
            v-for="request in store.filteredRequests"
            :key="request.request_id"
            class="rounded-3xl border border-white/10 bg-ink-900/70 p-4"
          >
            <div class="flex flex-wrap items-start justify-between gap-3">
              <div>
                <p class="text-sm font-semibold">{{ request.downstream }} -> {{ request.upstream || "-" }}</p>
                <p class="mt-1 text-xs text-slate-400">{{ request.request_id }}</p>
              </div>
              <div class="flex flex-wrap items-center gap-2">
                <code class="rounded-full bg-black/20 px-3 py-1 font-mono text-xs">{{ request.model || "-" }}</code>
                <span
                  class="rounded-full px-3 py-1 text-xs font-semibold"
                  :class="request.status_code >= 400 ? 'bg-signal-coral/15 text-signal-coral' : 'bg-signal-mint/15 text-signal-mint'"
                >
                  {{ request.status_code || "-" }}
                </span>
                <span class="rounded-full bg-white/5 px-3 py-1 text-xs text-slate-300">{{ request.duration_ms }} ms</span>
              </div>
            </div>
            <p class="mt-3 text-xs text-slate-500">{{ formatDateTime(request.created_at) }}</p>
            <p v-if="request.error" class="mt-3 rounded-2xl border border-signal-coral/15 bg-signal-coral/10 px-4 py-3 text-sm text-rose-100">
              {{ request.error }}
            </p>
          </article>
        </div>
      </PanelBlock>
    </div>
  </section>
</template>

<script setup>
import { onBeforeUnmount, onMounted, watch } from "vue";
import { useTrafficStore } from "../stores/traffic";

import PanelBlock from "../components/PanelBlock.vue";
import StatCard from "../components/StatCard.vue";

const store = useTrafficStore();
let refreshTimer = null;

function stopTimer() {
  if (refreshTimer) {
    clearInterval(refreshTimer);
    refreshTimer = null;
  }
}

function startTimer() {
  stopTimer();
  if (!store.autoRefresh) {
    return;
  }
  refreshTimer = setInterval(() => {
    store.refresh();
  }, 6000);
}

function formatDateTime(value) {
  if (!value) {
    return "not available";
  }
  return new Date(value).toLocaleString();
}

watch(
  () => store.autoRefresh,
  () => {
    startTimer();
  },
);

onMounted(() => {
  store.load();
  startTimer();
});

onBeforeUnmount(() => {
  stopTimer();
});
</script>
