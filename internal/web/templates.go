package web

const dashboardHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Beszel — Server Monitor</title>
<script src="https://cdn.jsdelivr.net/npm/chart.js@4.4.0/dist/chart.umd.min.js"></script>
<style>
  * { box-sizing: border-box; margin: 0; padding: 0; }
  body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background: #0f1117; color: #e6edf3; min-height: 100vh; }
  header { background: #161b22; border-bottom: 1px solid #30363d; padding: 1rem 2rem; display: flex; align-items: center; justify-content: space-between; }
  header h1 { font-size: 1.25rem; font-weight: 600; }
  header .hostname { color: #8b949e; font-size: 0.875rem; }
  main { padding: 1.5rem 2rem; max-width: 1200px; margin: 0 auto; }
  .gauge-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 1rem; margin-bottom: 1.5rem; }
  .gauge-card { background: #161b22; border: 1px solid #30363d; border-radius: 8px; padding: 1.25rem; }
  .gauge-card h3 { font-size: 0.75rem; text-transform: uppercase; letter-spacing: 0.05em; color: #8b949e; margin-bottom: 0.75rem; }
  .gauge-value { font-size: 2rem; font-weight: 700; }
  .gauge-detail { font-size: 0.8rem; color: #8b949e; margin-top: 0.25rem; }
  .chart-section { background: #161b22; border: 1px solid #30363d; border-radius: 8px; padding: 1.25rem; margin-bottom: 1.5rem; }
  .chart-section h3 { font-size: 0.875rem; margin-bottom: 1rem; }
  .chart-tabs { display: flex; gap: 0.5rem; margin-bottom: 1rem; }
  .chart-tabs button { background: #21262d; border: 1px solid #30363d; color: #8b949e; padding: 0.35rem 0.75rem; border-radius: 6px; cursor: pointer; font-size: 0.8rem; }
  .chart-tabs button.active { background: #388bfd33; border-color: #388bfd; color: #58a6ff; }
  .containers-section { background: #161b22; border: 1px solid #30363d; border-radius: 8px; padding: 1.25rem; }
  .containers-section h3 { font-size: 0.875rem; margin-bottom: 1rem; }
  .container-list { display: flex; flex-direction: column; gap: 0.5rem; }
  .container-item { background: #21262d; border-radius: 6px; padding: 0.75rem 1rem; display: flex; justify-content: space-between; align-items: center; }
  .container-name { font-size: 0.875rem; font-weight: 500; }
  .container-stats { display: flex; gap: 1.5rem; font-size: 0.8rem; color: #8b949e; }
  .updated { text-align: center; font-size: 0.75rem; color: #484f58; margin-top: 1rem; }
  @media (max-width: 600px) { main { padding: 1rem; } header { padding: 1rem; } }
</style>
</head>
<body>
<header>
  <h1>⚡ Beszel</h1>
  <span class="hostname" id="hostname">—</span>
</header>
<main>
  <div class="gauge-grid">
    <div class="gauge-card">
      <h3>CPU</h3>
      <div class="gauge-value" id="cpu-val" style="color:#58a6ff">—</div>
      <div class="gauge-detail" id="cpu-detail">—</div>
    </div>
    <div class="gauge-card">
      <h3>Memory</h3>
      <div class="gauge-value" id="mem-val" style="color:#a371f7">—</div>
      <div class="gauge-detail" id="mem-detail">—</div>
    </div>
    <div class="gauge-card">
      <h3>Disk</h3>
      <div class="gauge-value" id="disk-val" style="color:#3fb950">—</div>
      <div class="gauge-detail" id="disk-detail">—</div>
    </div>
    <div class="gauge-card">
      <h3>Network</h3>
      <div class="gauge-value" id="net-val" style="color:#f0883e">—</div>
      <div class="gauge-detail" id="net-detail">RX / TX</div>
    </div>
  </div>

  <div class="chart-section">
    <div class="chart-tabs">
      <button class="active" onclick="setChartRange(24)">24h</button>
      <button onclick="setChartRange(168)">7d</button>
    </div>
    <h3>System Metrics</h3>
    <canvas id="metricsChart" height="100"></canvas>
  </div>

  <div class="containers-section">
    <h3>🐳 Docker Containers</h3>
    <div class="container-list" id="container-list">
      <div style="color:#8b949e;font-size:0.875rem">No containers found</div>
    </div>
  </div>
  <div class="updated" id="updated">Loading…</div>
</main>
<script>
let chart = null;
let chartRange = 24;

function fmtBytes(b) {
  if (b < 1024) return b + ' B';
  if (b < 1024*1024) return (b/1024).toFixed(1) + ' KB';
  if (b < 1024*1024*1024) return (b/(1024*1024)).toFixed(1) + ' MB';
  return (b/(1024*1024*1024)).toFixed(2) + ' GB';
}

function colorFor(pct) {
  if (pct < 50) return '#3fb950';
  if (pct < 80) return '#f0883e';
  return '#f85149';
}

async function loadMetrics() {
  const res = await fetch('/api/metrics');
  if (!res.ok) return;
  const m = await res.json();
  document.getElementById('hostname').textContent = m.hostname || 'unknown';
  document.getElementById('cpu-val').textContent = m.cpu_percent.toFixed(1) + '%';
  document.getElementById('cpu-val').style.color = colorFor(m.cpu_percent);
  document.getElementById('mem-val').textContent = m.memory_percent.toFixed(1) + '%';
  document.getElementById('mem-val').style.color = colorFor(m.memory_percent);
  document.getElementById('mem-detail').textContent = fmtBytes(m.memory_used) + ' / ' + fmtBytes(m.memory_total);
  document.getElementById('disk-val').textContent = m.disk_percent.toFixed(1) + '%';
  document.getElementById('disk-val').style.color = colorFor(m.disk_percent);
  document.getElementById('disk-detail').textContent = fmtBytes(m.disk_used) + ' / ' + fmtBytes(m.disk_total);
  document.getElementById('net-val').textContent = fmtBytes(m.network_rx + m.network_tx);
  document.getElementById('net-detail').textContent = '↓ ' + fmtBytes(m.network_rx) + '  ↑ ' + fmtBytes(m.network_tx);
  document.getElementById('updated').textContent = 'Updated ' + new Date().toLocaleTimeString();
}

async function loadHistory() {
  const res = await fetch('/api/metrics/history?hours=' + chartRange);
  if (!res.ok) return;
  const data = await res.json();
  if (!data || data.length === 0) return;
  const labels = data.map(d => new Date(d.timestamp).toLocaleTimeString());
  const cpu = data.map(d => d.cpu_percent);
  const mem = data.map(d => d.memory_percent);
  const disk = data.map(d => d.disk_percent);
  if (chart) chart.destroy();
  const ctx = document.getElementById('metricsChart').getContext('2d');
  chart = new Chart(ctx, {
    type: 'line',
    data: {
      labels,
      datasets: [
        { label: 'CPU %', data: cpu, borderColor: '#58a6ff', backgroundColor: '#58a6ff22', tension: 0.3, fill: true },
        { label: 'Memory %', data: mem, borderColor: '#a371f7', backgroundColor: '#a371f722', tension: 0.3, fill: true },
        { label: 'Disk %', data: disk, borderColor: '#3fb950', backgroundColor: '#3fb95022', tension: 0.3, fill: true },
      ]
    },
    options: {
      responsive: true,
      plugins: { legend: { labels: { color: '#8b949e' } } },
      scales: {
        x: { ticks: { color: '#484f58', maxTicksLimit: 8 }, grid: { color: '#21262d' } },
        y: { min: 0, max: 100, ticks: { color: '#484f58' }, grid: { color: '#21262d' } }
      }
    }
  });
}

async function loadContainers() {
  const res = await fetch('/api/containers');
  if (!res.ok) return;
  const containers = await res.json();
  const el = document.getElementById('container-list');
  if (!containers || containers.length === 0) {
    el.innerHTML = '<div style="color:#8b949e;font-size:0.875rem">No containers found</div>';
    return;
  }
  el.innerHTML = containers.map(c => {
    const cpu = c.cpu_percent.toFixed(1);
    const mem = c.memory_percent.toFixed(1);
    return '<div class="container-item">' +
      '<span class="container-name">' + c.container_name + '</span>' +
      '<div class="container-stats">' +
        '<span>CPU: ' + cpu + '%</span>' +
        '<span>Mem: ' + mem + '%</span>' +
      '</div>' +
    '</div>';
  }).join('');
}

function setChartRange(hours) {
  chartRange = hours;
  document.querySelectorAll('.chart-tabs button').forEach(b => b.classList.remove('active'));
  event.target.classList.add('active');
  loadHistory();
}

async function init() {
  await loadMetrics();
  await loadHistory();
  await loadContainers();
  setInterval(loadMetrics, 5000);
  setInterval(loadHistory, 30000);
  setInterval(loadContainers, 15000);
}
init();
</script>
</body>
</html>`
