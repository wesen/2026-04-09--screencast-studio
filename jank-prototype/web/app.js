const appState = {
  videoSources: [],
};

const el = {
  dslInput: document.getElementById("dslInput"),
  applyBtn: document.getElementById("applyBtn"),
  startBtn: document.getElementById("startBtn"),
  stopBtn: document.getElementById("stopBtn"),
  loadExampleBtn: document.getElementById("loadExampleBtn"),
  clearBtn: document.getElementById("clearBtn"),
  configStatus: document.getElementById("configStatus"),
  recordingStatus: document.getElementById("recordingStatus"),
  sessionStatus: document.getElementById("sessionStatus"),
  startedStatus: document.getElementById("startedStatus"),
  outputList: document.getElementById("outputList"),
  previewGrid: document.getElementById("previewGrid"),
  previewCount: document.getElementById("previewCount"),
  warningList: document.getElementById("warningList"),
  logView: document.getElementById("logView"),
};

async function requestJSON(url, options = {}) {
  const resp = await fetch(url, options);
  if (!resp.ok) {
    const text = await resp.text();
    throw new Error(text || `${resp.status} ${resp.statusText}`);
  }
  return resp.json();
}

async function loadExample() {
  const resp = await fetch("/api/example");
  el.dslInput.value = await resp.text();
}

function escapeHTML(text) {
  return String(text)
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&#039;");
}

function renderPreviews(sources = []) {
  appState.videoSources = sources;
  el.previewGrid.innerHTML = "";
  el.previewCount.textContent = `${sources.length} source${sources.length === 1 ? "" : "s"}`;
  if (!sources.length) {
    el.previewGrid.classList.add("empty-state");
    el.previewGrid.textContent = "No enabled video sources.";
    return;
  }
  el.previewGrid.classList.remove("empty-state");
  for (const source of sources) {
    const card = document.createElement("article");
    card.className = "preview-card";
    card.innerHTML = `
      <div class="preview-card-header">
        <strong>${escapeHTML(source.name)}</strong>
        <span>${escapeHTML(source.type)}</span>
      </div>
      <img alt="${escapeHTML(source.name)} preview" src="${source.preview_url}?t=${Date.now()}" />
      <div class="preview-card-footer">${escapeHTML(source.id)}</div>
    `;
    el.previewGrid.appendChild(card);
  }
}

function renderOutputs(outputs = []) {
  el.outputList.innerHTML = "";
  if (!outputs.length) {
    const li = document.createElement("li");
    li.className = "muted";
    li.textContent = "No capture outputs yet.";
    el.outputList.appendChild(li);
    return;
  }
  for (const output of outputs) {
    const li = document.createElement("li");
    li.innerHTML = `<strong>${escapeHTML(output.name)}</strong><span>${escapeHTML(output.path)}</span>`;
    el.outputList.appendChild(li);
  }
}

function renderWarnings(warnings = []) {
  el.warningList.innerHTML = "";
  if (!warnings.length) {
    const li = document.createElement("li");
    li.textContent = "None";
    el.warningList.appendChild(li);
    return;
  }
  for (const warning of warnings) {
    const li = document.createElement("li");
    li.textContent = warning;
    el.warningList.appendChild(li);
  }
}

function renderLogs(logs = []) {
  el.logView.textContent = logs.join("\n");
  el.logView.scrollTop = el.logView.scrollHeight;
}

async function applyConfig() {
  const dsl = el.dslInput.value.trim();
  if (!dsl) {
    alert("Paste YAML first.");
    return;
  }
  el.applyBtn.disabled = true;
  try {
    const data = await requestJSON("/api/config/apply", {
      method: "POST",
      headers: { "Content-Type": "text/plain; charset=utf-8" },
      body: dsl,
    });
    el.configStatus.textContent = "Loaded";
    el.sessionStatus.textContent = data.session_id || "—";
    renderPreviews(data.video_sources || []);
    renderWarnings(data.warnings || []);
    await refreshState();
  } catch (err) {
    alert(err.message);
  } finally {
    el.applyBtn.disabled = false;
  }
}

async function startCapture() {
  el.startBtn.disabled = true;
  try {
    await requestJSON("/api/capture/start", { method: "POST" });
    await refreshState();
  } catch (err) {
    alert(err.message);
  } finally {
    el.startBtn.disabled = false;
  }
}

async function stopCapture() {
  el.stopBtn.disabled = true;
  try {
    await requestJSON("/api/capture/stop", { method: "POST" });
    await refreshState();
  } catch (err) {
    alert(err.message);
  } finally {
    el.stopBtn.disabled = false;
  }
}

async function refreshState() {
  try {
    const data = await requestJSON("/api/state");
    el.configStatus.textContent = data.configured ? "Loaded" : "None";
    el.recordingStatus.textContent = data.recording ? "Yes" : "No";
    el.sessionStatus.textContent = data.session_id || "—";
    el.startedStatus.textContent = data.started_at || "—";
    el.startBtn.disabled = !data.configured || data.recording;
    el.stopBtn.disabled = !data.recording;
    renderOutputs(data.outputs || []);
    renderWarnings(data.warnings || []);
    renderLogs(data.logs || []);

    if (data.configured) {
      const current = JSON.stringify(appState.videoSources || []);
      const next = JSON.stringify(data.video_sources || []);
      if (current !== next) {
        renderPreviews(data.video_sources || []);
      }
    }
  } catch (err) {
    console.error(err);
  }
}

function bind() {
  el.applyBtn.addEventListener("click", applyConfig);
  el.startBtn.addEventListener("click", startCapture);
  el.stopBtn.addEventListener("click", stopCapture);
  el.loadExampleBtn.addEventListener("click", loadExample);
  el.clearBtn.addEventListener("click", () => {
    el.dslInput.value = "";
  });
}

async function init() {
  bind();
  await loadExample();
  await refreshState();
  window.setInterval(refreshState, 1500);
}

init();
