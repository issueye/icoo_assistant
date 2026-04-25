import "./style.css";
import "./app.css";

import { GetOverview, ReloadProxy } from "../wailsjs/go/main/App";

const appRoot = document.querySelector("#app");

const state = {
  loading: true,
  refreshing: false,
  error: "",
  overview: null,
};

function escapeHtml(value) {
  return String(value ?? "")
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&#39;");
}

function renderBadge(active) {
  const label = active ? "Running" : "Stopped";
  const tone = active ? "ok" : "danger";
  return `<span class="badge badge-${tone}">${label}</span>`;
}

function renderRouteRows(items, emptyText) {
  if (!items || items.length === 0) {
    return `<div class="empty">${escapeHtml(emptyText)}</div>`;
  }
  return items
    .map(
      (item) => `
        <div class="row-item">
          <div>
            <div class="row-title">${escapeHtml(item.name)}</div>
            <div class="row-subtitle">${escapeHtml(item.upstream)}</div>
          </div>
          <code>${escapeHtml(item.model)}</code>
        </div>
      `,
    )
    .join("");
}

function renderUpstreams(items) {
  return (items || [])
    .map(
      (item) => `
        <article class="mini-card">
          <div class="mini-label">${escapeHtml(item.protocol)}</div>
          <div class="mini-value">${item.configured ? "Configured" : "Missing Key"}</div>
          <div class="mini-meta">${escapeHtml(item.base_url || "-")}</div>
        </article>
      `,
    )
    .join("");
}

function renderChecks(checks) {
  const entries = Object.entries(checks || {});
  if (entries.length === 0) {
    return `<div class="empty">No readiness checks yet.</div>`;
  }
  return entries
    .map(
      ([key, value]) => `
        <div class="check-pill ${value ? "check-pill-ok" : "check-pill-warn"}">
          <span>${escapeHtml(key)}</span>
          <strong>${escapeHtml(String(value))}</strong>
        </div>
      `,
    )
    .join("");
}

function renderNotes(notes) {
  return (notes || [])
    .map((note) => `<li>${escapeHtml(note)}</li>`)
    .join("");
}

function renderPaths(paths) {
  return (paths || [])
    .map((path) => `<code class="path-chip">${escapeHtml(path)}</code>`)
    .join("");
}

function renderRecentRequests(items) {
  if (!items || items.length === 0) {
    return `<div class="empty">No requests recorded yet.</div>`;
  }
  return items
    .map(
      (item) => `
        <div class="request-row">
          <div class="request-main">
            <div class="request-title">${escapeHtml(item.downstream)} -> ${escapeHtml(item.upstream || "-")}</div>
            <div class="request-meta">${escapeHtml(item.request_id)} | ${escapeHtml(item.created_at)}</div>
          </div>
          <div class="request-side">
            <code>${escapeHtml(item.model || "-")}</code>
            <span class="request-status ${item.status_code >= 400 ? "request-status-error" : "request-status-ok"}">
              ${escapeHtml(String(item.status_code))}
            </span>
            <span class="request-duration">${escapeHtml(String(item.duration_ms))} ms</span>
          </div>
          ${item.error ? `<div class="request-error">${escapeHtml(item.error)}</div>` : ""}
        </div>
      `,
    )
    .join("");
}

function render() {
  if (state.loading) {
    appRoot.innerHTML = `
      <main class="shell">
        <section class="hero hero-loading">
          <div class="eyebrow">icoo_proxy</div>
          <h1>Starting desktop gateway console...</h1>
          <p>Loading local proxy state, model routes, and upstream availability.</p>
        </section>
      </main>
    `;
    return;
  }

  const overview = state.overview || {};
  const authMode = overview.auth_required ? "API key required" : "Local trusted mode";
  const danger = state.error || overview.last_error;

  appRoot.innerHTML = `
    <main class="shell">
      <section class="hero">
        <div class="hero-copy">
          <div class="eyebrow">Local AI Gateway</div>
          <h1>One desktop entrypoint for Anthropic, Chat, and Responses.</h1>
          <p class="hero-text">
            The Wails shell keeps the local proxy visible and manageable while the Go backend handles
            routing, auth, and protocol-compatible forwarding.
          </p>
          <div class="hero-actions">
            <button class="primary-btn" id="reload-proxy" ${state.refreshing ? "disabled" : ""}>
              ${state.refreshing ? "Reloading..." : "Reload Proxy"}
            </button>
            ${renderBadge(Boolean(overview.running))}
          </div>
        </div>
        <div class="hero-panel">
          <div class="stat-card">
            <span class="stat-label">Proxy URL</span>
            <strong class="stat-value">${escapeHtml(overview.proxy_url || "Not running")}</strong>
          </div>
          <div class="stat-card">
            <span class="stat-label">Listen Address</span>
            <strong class="stat-value">${escapeHtml(overview.listen_addr || "-")}</strong>
          </div>
          <div class="stat-card">
            <span class="stat-label">Access Mode</span>
            <strong class="stat-value">${escapeHtml(authMode)}</strong>
          </div>
          <div class="stat-card">
            <span class="stat-label">Version</span>
            <strong class="stat-value">${escapeHtml(overview.version || "-")}</strong>
          </div>
        </div>
      </section>

      ${danger ? `<section class="alert">${escapeHtml(danger)}</section>` : ""}

      <section class="grid two-up">
        <article class="panel">
          <div class="panel-header">
            <div>
              <div class="panel-eyebrow">Upstreams</div>
              <h2>Provider readiness</h2>
            </div>
          </div>
          <div class="mini-grid">${renderUpstreams(overview.upstreams)}</div>
          <div class="checks">${renderChecks(overview.checks)}</div>
        </article>

        <article class="panel">
          <div class="panel-header">
            <div>
              <div class="panel-eyebrow">Paths</div>
              <h2>Supported endpoints</h2>
            </div>
          </div>
          <div class="path-grid">${renderPaths(overview.supported_paths)}</div>
        </article>
      </section>

      <section class="grid two-up">
        <article class="panel">
          <div class="panel-header">
            <div>
              <div class="panel-eyebrow">Defaults</div>
              <h2>Default route targets</h2>
            </div>
          </div>
          <div class="stack">${renderRouteRows(overview.defaults, "No default routes configured yet.")}</div>
        </article>

        <article class="panel">
          <div class="panel-header">
            <div>
              <div class="panel-eyebrow">Aliases</div>
              <h2>Model alias catalog</h2>
            </div>
          </div>
          <div class="stack">${renderRouteRows(overview.aliases, "No aliases configured yet.")}</div>
        </article>
      </section>

      <section class="grid single">
        <article class="panel">
          <div class="panel-header">
            <div>
              <div class="panel-eyebrow">Traffic</div>
              <h2>Recent requests</h2>
            </div>
          </div>
          <div class="stack">${renderRecentRequests(overview.recent_requests)}</div>
        </article>
      </section>

      <section class="grid single">
        <article class="panel panel-notes">
          <div class="panel-header">
            <div>
              <div class="panel-eyebrow">Build notes</div>
              <h2>Current implementation scope</h2>
            </div>
          </div>
          <ul class="notes-list">${renderNotes(overview.notes)}</ul>
        </article>
      </section>
    </main>
  `;

  const reloadButton = document.querySelector("#reload-proxy");
  if (reloadButton) {
    reloadButton.addEventListener("click", refreshProxy);
  }
}

async function loadOverview() {
  state.loading = true;
  state.error = "";
  render();

  try {
    state.overview = await GetOverview();
  } catch (error) {
    state.error = error?.message || String(error);
  } finally {
    state.loading = false;
    render();
  }
}

async function refreshProxy() {
  state.refreshing = true;
  state.error = "";
  render();

  try {
    state.overview = await ReloadProxy();
  } catch (error) {
    state.error = error?.message || String(error);
  } finally {
    state.refreshing = false;
    render();
  }
}

loadOverview();
