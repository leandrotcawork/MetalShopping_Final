#!/usr/bin/env node

const fs = require('fs');
const path = require('path');
const sqlite3 = require('sqlite3').verbose();

const colors = {
  cyan: '\x1b[36m',
  green: '\x1b[32m',
  yellow: '\x1b[33m',
  blue: '\x1b[34m',
  reset: '\x1b[0m',
};

function log(color, text) {
  console.log(`${color}${text}${colors.reset}`);
}

function generateHTML(sinapses, links) {
  const regionColors = {
    'hippocampus': '#FFD700',
    'cortex/backend': '#4169E1',
    'cortex/frontend': '#9370DB',
    'cortex/database': '#FF8C00',
    'cortex/infra': '#808080',
    'sinapses': '#20B2AA',
    'lessons': '#DC143C',
  };

  const nodes = sinapses.map(s => ({
    id: s.id,
    label: s.title,
    region: s.region,
    weight: s.weight || 0.5,
    severity: s.severity || 'medium',
    color: regionColors[s.region] || '#999999',
  }));

  const nodesByID = new Map(nodes.map(n => [n.id, n]));
  const d3links = links
    .filter(l => nodesByID.has(l.source_id) && nodesByID.has(l.target_id))
    .map(l => ({
      source: l.source_id,
      target: l.target_id,
    }));

  const nodesJSON = JSON.stringify(nodes);
  const linksJSON = JSON.stringify(d3links);

  return `<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <title>MetalShopping Brain Graph</title>
  <style>
    body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; margin: 0; padding: 20px; background: #1a1a1a; color: #eee; }
    #container { position: relative; width: 100%; height: 90vh; border: 1px solid #333; background: #0d0d0d; }
    svg { width: 100%; height: 100%; }
    .node { cursor: pointer; stroke: #fff; stroke-width: 1.5px; }
    .node:hover { stroke-width: 3px; }
    .link { stroke: #999; stroke-opacity: 0.3; stroke-width: 1px; }
    .label { font-size: 12px; pointer-events: none; text-anchor: middle; fill: #eee; }
    #info { position: absolute; top: 20px; right: 20px; width: 300px; background: #222; border: 1px solid #444; border-radius: 4px; padding: 15px; max-height: 400px; overflow-y: auto; }
    #info h3 { margin-top: 0; color: #FFD700; }
    #info p { margin: 5px 0; font-size: 12px; }
    #legend { position: absolute; bottom: 20px; left: 20px; background: #222; border: 1px solid #444; border-radius: 4px; padding: 15px; }
    .legend-item { display: flex; align-items: center; margin: 5px 0; font-size: 12px; }
    .legend-color { width: 20px; height: 20px; margin-right: 10px; border-radius: 50%; }
    .controls { position: absolute; top: 20px; left: 20px; background: #222; border: 1px solid #444; border-radius: 4px; padding: 10px; }
    button { background: #4169E1; color: #fff; border: none; padding: 8px 15px; border-radius: 4px; cursor: pointer; margin: 5px; }
    button:hover { background: #5A7FDB; }
  </style>
</head>
<body>
  <div id="container"></div>
  <div id="info" style="display:none;"></div>
  <div id="legend">
    <div class="legend-item"><div class="legend-color" style="background:#FFD700;"></div>Hippocampus</div>
    <div class="legend-item"><div class="legend-color" style="background:#4169E1;"></div>Backend</div>
    <div class="legend-item"><div class="legend-color" style="background:#9370DB;"></div>Frontend</div>
    <div class="legend-item"><div class="legend-color" style="background:#FF8C00;"></div>Database</div>
    <div class="legend-item"><div class="legend-color" style="background:#808080;"></div>Infra</div>
    <div class="legend-item"><div class="legend-color" style="background:#20B2AA;"></div>Sinapses</div>
    <div class="legend-item"><div class="legend-color" style="background:#DC143C;"></div>Lessons</div>
  </div>
  <div class="controls">
    <button onclick="resetZoom()">Reset View</button>
    <button onclick="toggleLabels()">Toggle Labels</button>
  </div>

  <script src="https://d3js.org/d3.v7.min.js"></script>
  <script>
    const nodes = ${nodesJSON};
    const links = ${linksJSON};
    let showLabels = true;

    const width = document.getElementById('container').clientWidth;
    const height = document.getElementById('container').clientHeight;

    const svg = d3.select('#container')
      .append('svg')
      .attr('width', width)
      .attr('height', height);

    const g = svg.append('g');

    const simulation = d3.forceSimulation(nodes)
      .force('link', d3.forceLink(links)
        .id(d => d.id)
        .distance(100))
      .force('charge', d3.forceManyBody().strength(-300))
      .force('center', d3.forceCenter(width / 2, height / 2))
      .force('collision', d3.forceCollide().radius(30));

    const link = g.append('g')
      .selectAll('line')
      .data(links)
      .enter()
      .append('line')
      .attr('class', 'link');

    const node = g.append('g')
      .selectAll('circle')
      .data(nodes)
      .enter()
      .append('circle')
      .attr('class', 'node')
      .attr('r', d => 8 + (d.weight || 0.5) * 5)
      .attr('fill', d => d.color)
      .call(d3.drag()
        .on('start', dragstarted)
        .on('drag', dragged)
        .on('end', dragended))
      .on('click', showInfo);

    const label = g.append('g')
      .selectAll('text')
      .data(nodes)
      .enter()
      .append('text')
      .attr('class', 'label')
      .text(d => d.label)
      .attr('dy', '0.3em');

    const zoom = d3.zoom()
      .on('zoom', (event) => {
        g.attr('transform', event.transform);
      });

    svg.call(zoom);

    simulation.on('tick', () => {
      link
        .attr('x1', d => d.source.x)
        .attr('y1', d => d.source.y)
        .attr('x2', d => d.target.x)
        .attr('y2', d => d.target.y);

      node
        .attr('cx', d => d.x)
        .attr('cy', d => d.y);

      label
        .attr('x', d => d.x)
        .attr('y', d => d.y)
        .style('display', showLabels ? 'block' : 'none');
    });

    function dragstarted(event, d) {
      if (!event.active) simulation.alphaTarget(0.3).restart();
      d.fx = d.x;
      d.fy = d.y;
    }

    function dragged(event, d) {
      d.fx = event.x;
      d.fy = event.y;
    }

    function dragended(event, d) {
      if (!event.active) simulation.alphaTarget(0);
      d.fx = null;
      d.fy = null;
    }

    function showInfo(event, d) {
      const infoDiv = document.getElementById('info');
      const infoContent = document.createElement('div');
      const title = document.createElement('h3');
      title.textContent = d.label;
      infoContent.appendChild(title);

      const id = document.createElement('p');
      id.innerHTML = '<strong>ID:</strong> ' + escapeHtml(d.id);
      infoContent.appendChild(id);

      const region = document.createElement('p');
      region.innerHTML = '<strong>Region:</strong> ' + escapeHtml(d.region);
      infoContent.appendChild(region);

      const weight = document.createElement('p');
      weight.innerHTML = '<strong>Weight:</strong> ' + (d.weight || 0.5).toFixed(2);
      infoContent.appendChild(weight);

      if (d.severity) {
        const severity = document.createElement('p');
        severity.innerHTML = '<strong>Severity:</strong> ' + escapeHtml(d.severity);
        infoContent.appendChild(severity);
      }

      infoDiv.textContent = '';
      infoDiv.appendChild(infoContent);
      infoDiv.style.display = 'block';
    }

    function escapeHtml(text) {
      const map = {
        '&': '&amp;',
        '<': '&lt;',
        '>': '&gt;',
        '"': '&quot;',
        "'": '&#039;'
      };
      return text.replace(/[&<>"']/g, m => map[m]);
    }

    function resetZoom() {
      svg.transition().duration(750).call(
        zoom.transform,
        d3.zoomIdentity.translate(0, 0).scale(1)
      );
    }

    function toggleLabels() {
      showLabels = !showLabels;
      label.style('display', showLabels ? 'block' : 'none');
    }
  </script>
</body>
</html>`;
}

function main() {
  log(colors.cyan, '\n Brain Visualizer\n');

  const projectPath = process.cwd();
  const brainRoot = path.join(projectPath, '.brain');
  const dbPath = path.join(brainRoot, 'brain.db');

  if (!fs.existsSync(dbPath)) {
    log(colors.yellow, 'brain.db not found. Run: python build_brain_db.py');
    process.exit(1);
  }

  const db = new sqlite3.Database(dbPath);

  db.all('SELECT id, title, region, weight, severity FROM sinapses', (err, sinapses) => {
    if (err) {
      log(colors.yellow, `Error: ${err.message}`);
      process.exit(1);
    }

    db.all('SELECT source_id, target_id FROM sinapse_links', (err, links) => {
      if (err) {
        log(colors.yellow, `Error: ${err.message}`);
        process.exit(1);
      }

      const htmlPath = path.join(brainRoot, 'brain-graph.html');
      const html = generateHTML(sinapses, links);

      fs.writeFileSync(htmlPath, html, 'utf-8');

      log(colors.green, `Visualization created!`);
      log(colors.blue, `   HTML: ${htmlPath}`);
      log(colors.blue, `   Sinapses: ${sinapses.length}`);
      log(colors.blue, `   Links: ${links.length}\n`);

      db.close();
    });
  });
}

main();
