/* fresnel-tm console: every number on the page comes from the backend. */
"use strict";

const $ = (id) => document.getElementById(id);

function showError(msg) {
  const panel = $("error-panel");
  panel.hidden = false;
  $("error-body").textContent = msg;
}

function clearError() {
  $("error-panel").hidden = true;
}

function renderLayers(layers) {
  const body = $("layers-body");
  body.innerHTML = "";
  if (!layers || layers.length === 0) {
    body.innerHTML = '<tr><td colspan="4" style="text-align:center">无镀层（裸界面）</td></tr>';
    return;
  }
  layers.forEach((l) => appendLayerRow(l.index, l.extinction || 0, l.thickness_nm));
}

function appendLayerRow(index, extinction, thickness) {
  const body = $("layers-body");
  if (body.querySelector(".empty-row")) body.innerHTML = "";
  const tr = document.createElement("tr");
  tr.innerHTML = `
    <td><input type="number" step="0.01" min="0" class="n" value="${index}"></td>
    <td><input type="number" step="0.01" min="0" class="k" value="${extinction}"></td>
    <td><input type="number" step="0.01" min="0" class="d" value="${thickness}"></td>
    <td><button type="button" class="del">删</button></td>`;
  tr.querySelector(".del").addEventListener("click", () => tr.remove());
  body.appendChild(tr);
}

function readStackRequest() {
  const layers = [];
  document.querySelectorAll("#layers-body tr").forEach((tr) => {
    if (tr.classList.contains("empty-row")) return;
    const n = parseFloat(tr.querySelector(".n").value);
    const k = parseFloat(tr.querySelector(".k").value) || 0;
    const d = parseFloat(tr.querySelector(".d").value);
    if (Number.isFinite(n) && Number.isFinite(d)) {
      layers.push({ index: n, thickness_nm: d, extinction: k });
    }
  });
  return {
    incident: { index: parseFloat($("incident-index").value) },
    layers: layers,
    substrate: { index: parseFloat($("substrate-index").value) },
    wavelength_nm: parseFloat($("wavelength").value),
    angle_deg: parseFloat($("angle").value),
    polarization: $("polarization").value,
  };
}

async function postJSON(url, body) {
  const resp = await fetch(url, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(body),
  });
  const data = await resp.json().catch(() => ({ error: "后端返回的不是 JSON" }));
  if (!resp.ok) {
    throw new Error(data.error || "请求失败（HTTP " + resp.status + "）");
  }
  return data;
}

function fmt(x) {
  return (x * 100).toFixed(3) + "%";
}

function renderResult(res) {
  const rows = [
    ["反射率 R", fmt(res.reflection)],
    ["透射率 T", fmt(res.transmission)],
    ["吸收率 A", fmt(res.absorption)],
    ["能量和 R+A+T", res.energy_sum.toFixed(9)],
    ["裸基板反射率（参考）", fmt(res.bare_reflection)],
    ["镀膜减反量", fmt(res.bare_reflection - res.reflection)],
  ];
  $("result-table").innerHTML =
    "<table><tbody>" +
    rows.map(([k, v]) => `<tr><td>${k}</td><td>${v}</td></tr>`).join("") +
    "</tbody></table>";
  $("result-extra").textContent =
    "波长 " + res.wavelength_nm + " nm · 入射角 " + res.angle_deg + "° · 偏振 " + res.polarization;
}

function renderSpectrum(res) {
  const wrap = $("chart-wrap");
  const W = 920, H = 300, PAD = { l: 46, r: 16, t: 16, b: 30 };
  const pts = res.points;
  if (!pts || pts.length < 2) {
    wrap.innerHTML = '<p class="hint">后端没有返回光谱点。</p>';
    return;
  }
  const xs = pts.map((p) => p.wavelength_nm);
  const lo = res.reflection_range.min, hi = res.reflection_range.max;
  const pad = Math.max((hi - lo) * 0.08, 0.005);
  const yLo = Math.max(0, lo - pad), yHi = Math.min(1, hi + pad);
  const x = (wl) => PAD.l + ((wl - xs[0]) / (xs[xs.length - 1] - xs[0])) * (W - PAD.l - PAD.r);
  const y = (v) => H - PAD.b - ((v - yLo) / (yHi - yLo)) * (H - PAD.t - PAD.b);

  const series = [
    { key: "reflection", color: "#1a6fb5", label: "R(λ)" },
    { key: "transmission", color: "#2e8b57", label: "T(λ)" },
    { key: "absorption", color: "#b33a2e", label: "A(λ)" },
  ];
  const poly = (key) =>
    pts.map((p, i) => (i === 0 ? "M" : "L") + x(p.wavelength_nm).toFixed(1) + "," + y(p[key]).toFixed(1)).join(" ");
  const min = res.min_reflection;
  let mark = "";
  if (min && min.value < 1) {
    mark = `<circle cx="${x(min.wavelength_nm)}" cy="${y(min.value)}" r="3.5" fill="#d97706"/>
            <text x="${x(min.wavelength_nm) - 6}" y="${y(min.value) - 8}" text-anchor="end">R_min ${(min.value * 100).toFixed(2)}% @ ${min.wavelength_nm}nm</text>`;
  }

  let svg = `<svg viewBox="0 0 ${W} ${H}" width="100%" height="${H}" preserveAspectRatio="xMidYMid meet">`;
  for (let i = 0; i <= 4; i++) {
    const v = yLo + ((yHi - yLo) * i) / 4;
    svg += `<line x1="${PAD.l}" y1="${y(v)}" x2="${W - PAD.r}" y2="${y(v)}" stroke="#eef2f6"/>
            <text x="${PAD.l - 6}" y="${y(v) + 3}" text-anchor="end">${(v * 100).toFixed(0)}%</text>`;
  }
  series.forEach((s) => {
    svg += `<path d="${poly(s.key)}" fill="none" stroke="${s.color}" stroke-width="2"/>`;
  });
  svg += mark;
  svg += `<text x="${PAD.l}" y="${H - 8}">${xs[0]}nm</text>
          <text x="${W - PAD.r}" y="${H - 8}" text-anchor="end">${xs[xs.length - 1]}nm</text>`;
  svg += `<g transform="translate(${W - PAD.r - 130}, ${PAD.t})" font-size="12">` +
    series.map((s, i) => `<text x="0" y="${i * 18}"><tspan fill="${s.color}">■</tspan> ${s.label}</text>`).join("") +
    "</g>";
  svg += "</svg>";
  wrap.innerHTML = svg;
}

async function initExamples() {
  try {
    const resp = await fetch("/api/examples");
    const data = await resp.json();
    const box = $("example-buttons");
    box.innerHTML = "";
    (data.examples || []).forEach((ex) => {
      const b = document.createElement("button");
      b.textContent = ex.name;
      b.addEventListener("click", () => loadExample(ex.name));
      box.appendChild(b);
    });
  } catch (err) {
    showError("无法加载示例列表：" + err.message);
  }
}

async function loadExample(name) {
  clearError();
  try {
    const resp = await fetch("/api/examples/" + name);
    const ex = await resp.json();
    if (!resp.ok) throw new Error(ex.error || "加载示例失败");
    $("incident-index").value = ex.incident.index;
    $("substrate-index").value = ex.substrate.index;
    $("angle").value = ex.angle_deg || 0;
    $("wavelength").value = ex.wavelength_nm || ex.design_wavelength_nm || 550;
    $("polarization").value = ex.polarization || "average";
    renderLayers(ex.layers);
    $("example-desc").textContent = ex.description || "";
    $("wl-min").value = Math.round((ex.design_wavelength_nm || 550) * 0.7);
    $("wl-max").value = Math.round((ex.design_wavelength_nm || 550) * 1.3);
    await calc();
    await scan();
  } catch (err) {
    showError(err.message);
  }
}

async function calc() {
  clearError();
  try {
    const res = await postJSON("/api/stack", readStackRequest());
    renderResult(res);
  } catch (err) {
    showError("计算失败：" + err.message);
  }
}

async function scan() {
  clearError();
  const req = readStackRequest();
  req.wavelength_nm = undefined;
  delete req.wavelength_nm;
  req.wavelength_min_nm = parseFloat($("wl-min").value);
  req.wavelength_max_nm = parseFloat($("wl-max").value);
  req.points = parseInt($("wl-points").value, 10) || 201;
  try {
    const res = await postJSON("/api/spectrum", req);
    renderSpectrum(res);
  } catch (err) {
    showError("光谱扫描失败：" + err.message);
  }
}

document.addEventListener("DOMContentLoaded", () => {
  $("add-layer").addEventListener("click", () => appendLayerRow(1.38, 0, 100));
  $("calc").addEventListener("click", calc);
  $("scan").addEventListener("click", scan);
  renderLayers([{ index: 1.38, thickness_nm: 99.64 }]);
  initExamples();
  calc();
  scan();
});
