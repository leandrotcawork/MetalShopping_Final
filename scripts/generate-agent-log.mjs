#!/usr/bin/env node
// Agent Activity Log → HTML Dashboard
// Usage: node scripts/generate-agent-log.mjs
// Output: logs/agent-activity.html

import { readFileSync, writeFileSync, readdirSync } from 'fs';
import { join } from 'path';

const LOGS_DIR = 'logs';
const OUTPUT = join(LOGS_DIR, 'agent-activity.html');

const STATUS_COLOR = {
  success: '#22c55e',
  'fix-applied': '#eab308',
  escalated: '#f97316',
  failed: '#ef4444',
  'false-positive': '#94a3b8',
  default: '#94a3b8',
};

function readAllEntries() {
  const files = readdirSync(LOGS_DIR)
    .filter(f => f.endsWith('.jsonl') && f.startsWith('agent-activity-'));
  const entries = [];
  for (const file of files) {
    const lines = readFileSync(join(LOGS_DIR, file), 'utf-8')
      .split('\n')
      .filter(Boolean);
    for (const line of lines) {
      try { entries.push(JSON.parse(line)); } catch {}
    }
  }
  return entries.sort((a, b) => new Date(b.timestamp) - new Date(a.timestamp));
}

function buildHTML(entries) {
  const agents = [...new Set(entries.map(e => e.agent))].sort();
  const stages = [...new Set(entries.map(e => e.stage))].sort();
  const statuses = [...new Set(entries.map(e => e.status))].sort();

  const rows = entries.map(e => {
    const color = STATUS_COLOR[e.status] || STATUS_COLOR.default;
    const files = (e.files_changed || []).join(', ') || e.output_summary || '—';
    const commit = e.commit ? `<a href="#">${e.commit.slice(0, 7)}</a>` : '—';
    return `
      <tr data-agent="${e.agent}" data-stage="${e.stage}" data-status="${e.status}">
        <td>${new Date(e.timestamp).toLocaleString()}</td>
        <td><code>${e.agent}</code></td>
        <td>${e.stage}</td>
        <td>${e.task}</td>
        <td style="color:${color};font-weight:bold">${e.status}</td>
        <td>${commit}</td>
        <td class="expandable" title="${(e.decision||'').replace(/"/g,'&quot;')}">${(e.decision||'—').slice(0,60)}${(e.decision||'').length>60?'…':''}</td>
      </tr>`;
  }).join('');

  const agentOptions = agents.map(a => `<option value="${a}">${a}</option>`).join('');
  const stageOptions = stages.map(s => `<option value="${s}">${s}</option>`).join('');
  const statusOptions = statuses.map(s => `<option value="${s}">${s}</option>`).join('');

  return `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<title>Agent Activity Log — MetalShopping</title>
<style>
  body { font-family: -apple-system, sans-serif; background: #0f172a; color: #e2e8f0; margin: 0; padding: 24px; }
  h1 { color: #f8fafc; margin-bottom: 8px; }
  .meta { color: #94a3b8; font-size: 13px; margin-bottom: 24px; }
  .filters { display: flex; gap: 12px; margin-bottom: 16px; flex-wrap: wrap; }
  select, input { background: #1e293b; color: #e2e8f0; border: 1px solid #334155; padding: 6px 10px; border-radius: 6px; font-size: 13px; }
  input { width: 240px; }
  table { width: 100%; border-collapse: collapse; font-size: 13px; }
  th { background: #1e293b; color: #94a3b8; text-align: left; padding: 10px 12px; border-bottom: 1px solid #334155; position: sticky; top: 0; }
  td { padding: 9px 12px; border-bottom: 1px solid #1e293b; vertical-align: top; }
  tr:hover td { background: #1e293b; }
  code { background: #334155; padding: 2px 6px; border-radius: 4px; font-size: 12px; }
  a { color: #38bdf8; }
  .hidden { display: none; }
  .count { color: #94a3b8; font-size: 13px; margin-bottom: 8px; }
</style>
</head>
<body>
<h1>Agent Activity Log</h1>
<div class="meta">Generated ${new Date().toLocaleString()} · ${entries.length} entries</div>
<div class="filters">
  <input id="search" placeholder="Search tasks, agents, decisions…" oninput="filter()">
  <select id="agentFilter" onchange="filter()"><option value="">All agents</option>${agentOptions}</select>
  <select id="stageFilter" onchange="filter()"><option value="">All stages</option>${stageOptions}</select>
  <select id="statusFilter" onchange="filter()"><option value="">All statuses</option>${statusOptions}</select>
</div>
<div class="count" id="count">${entries.length} entries</div>
<table>
  <thead><tr>
    <th>Time</th><th>Agent</th><th>Stage</th><th>Task</th>
    <th>Status</th><th>Commit</th><th>Decision</th>
  </tr></thead>
  <tbody id="tbody">${rows}</tbody>
</table>
<script>
function filter() {
  const search = document.getElementById('search').value.toLowerCase();
  const agent = document.getElementById('agentFilter').value;
  const stage = document.getElementById('stageFilter').value;
  const status = document.getElementById('statusFilter').value;
  let visible = 0;
  document.querySelectorAll('#tbody tr').forEach(row => {
    const text = row.textContent.toLowerCase();
    const show = (!search || text.includes(search))
      && (!agent || row.dataset.agent === agent)
      && (!stage || row.dataset.stage === stage)
      && (!status || row.dataset.status === status);
    row.classList.toggle('hidden', !show);
    if (show) visible++;
  });
  document.getElementById('count').textContent = visible + ' entries';
}
</script>
</body>
</html>`;
}

const entries = readAllEntries();
writeFileSync(OUTPUT, buildHTML(entries));
console.log(`Generated ${OUTPUT} with ${entries.length} entries.`);
