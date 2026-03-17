#!/usr/bin/env node
// graph.mjs — Renders benchmark results.csv as SVG chart images
// Usage: node benchmark/graph.mjs [--light]
// Output: benchmark/graphs/*.svg
//
// Zero dependencies — pure Node.js SVG generation.
// SVGs have transparent backgrounds for use on any site.

import { readFileSync, writeFileSync, mkdirSync } from 'node:fs';
import { resolve } from 'node:path';

const CSV_PATH = resolve('benchmark/results.csv');
const OUT_DIR = resolve('benchmark/graphs');

// ── Theme ──────────────────────────────────────────────────────────────
const THEMES = {
  dark:  { bg: '#000000', text: '#f5f5f5', muted: '#a3a3a3', grid: '#2a2a2a', subtle: '#333333' },
  light: { bg: '#ffffff', text: '#1f2937', muted: '#6b7280', grid: '#d1d5db', subtle: '#9ca3af' },
};
const theme = process.argv.includes('--light') ? 'light' : 'dark';
const T = THEMES[theme];

const PHASE_COLORS = { baseline: '#ef4444', 'opt-v1': '#eab308', argon2id: '#22c55e' };
const METRIC_COLORS = { p50: '#3b82f6', p95: '#eab308', p99: '#ef4444' };

// ── CSV ────────────────────────────────────────────────────────────────
function readCsv() {
  const lines = readFileSync(CSV_PATH, 'utf8').trim().split('\n').slice(1);
  return lines.filter(l => l.trim()).map(line => {
    const [ts, scenario, profile, label, reqs, rps, avg, p50, p95, p99, errs] = line.split(',');
    return {
      timestamp: ts, scenario, profile, label,
      total_reqs: +reqs, rps: +rps, duration_avg: +avg,
      duration_p50: +p50, duration_p95: +p95, duration_p99: +p99, errors: +errs,
    };
  });
}

function getVal(rows, label, field) {
  const row = rows.find(r => r.label === label);
  return row ? row[field] : 0;
}

// ── SVG Helpers ────────────────────────────────────────────────────────
const esc = s => String(s).replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');

function niceMax(val) {
  if (val <= 0) return 10;
  const mag = Math.pow(10, Math.floor(Math.log10(val)));
  const norm = val / mag;
  if (norm <= 1) return mag;
  if (norm <= 2) return 2 * mag;
  if (norm <= 5) return 5 * mag;
  return 10 * mag;
}

function formatNum(n) {
  if (n >= 1000000) return (n / 1000000).toFixed(1) + 'M';
  if (n >= 100000) return Math.round(n / 1000) + 'K';
  if (n >= 1000) return (n / 1000).toFixed(1) + 'K';
  if (n >= 100) return Math.round(n).toString();
  if (n >= 10) return n.toFixed(1);
  return n.toFixed(2);
}

// ── Bar Chart ──────────────────────────────────────────────────────────
function svgBarChart({ title, categories, series, width = 800, height = 450, yLabel = '' }) {
  const m = { top: 60, right: 30, bottom: 90, left: 80 };
  const cw = width - m.left - m.right;
  const ch = height - m.top - m.bottom;

  const allVals = series.flatMap(s => s.values);
  const yMax = niceMax(Math.max(...allVals) * 1.12);
  const yScale = ch / yMax;

  const groupCount = categories.length;
  const barCount = series.length;
  const groupWidth = cw / groupCount;
  const groupPad = groupWidth * 0.15;
  const barWidth = (groupWidth - groupPad * 2) / barCount;
  const barGap = Math.max(1, barWidth * 0.06);

  let svg = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 ${width} ${height}" font-family="'Inter','Segoe UI',system-ui,-apple-system,sans-serif">\n`;
  svg += `  <rect width="${width}" height="${height}" fill="${T.bg}"/>\n`;

  // Title
  svg += `  <text x="${width / 2}" y="34" text-anchor="middle" fill="${T.text}" font-size="18" font-weight="700">${esc(title)}</text>\n`;

  // Y-axis label
  if (yLabel) {
    svg += `  <text x="${m.left - 55}" y="${m.top + ch / 2}" text-anchor="middle" fill="${T.muted}" font-size="12" transform="rotate(-90,${m.left - 55},${m.top + ch / 2})">${esc(yLabel)}</text>\n`;
  }

  // Gridlines + Y labels
  const gridSteps = 5;
  for (let i = 0; i <= gridSteps; i++) {
    const val = (yMax / gridSteps) * i;
    const y = m.top + ch - val * yScale;
    svg += `  <line x1="${m.left}" y1="${y.toFixed(1)}" x2="${m.left + cw}" y2="${y.toFixed(1)}" stroke="${T.grid}" stroke-width="0.5"${i > 0 ? ' stroke-dasharray="4,4"' : ''}/>\n`;
    svg += `  <text x="${m.left - 10}" y="${(y + 4).toFixed(1)}" text-anchor="end" fill="${T.muted}" font-size="11">${formatNum(val)}</text>\n`;
  }

  // Bars
  for (let gi = 0; gi < groupCount; gi++) {
    const groupX = m.left + gi * groupWidth + groupPad;

    for (let si = 0; si < barCount; si++) {
      const val = series[si].values[gi];
      const barH = Math.max(1, val * yScale);
      const x = groupX + si * barWidth + barGap / 2;
      const y = m.top + ch - barH;
      const w = barWidth - barGap;

      svg += `  <rect x="${x.toFixed(1)}" y="${y.toFixed(1)}" width="${w.toFixed(1)}" height="${barH.toFixed(1)}" fill="${series[si].color}" rx="3" opacity="0.9"/>\n`;

      // Value on top
      const fontSize = barCount > 3 ? 9 : 11;
      svg += `  <text x="${(x + w / 2).toFixed(1)}" y="${(y - 6).toFixed(1)}" text-anchor="middle" fill="${T.text}" font-size="${fontSize}" font-weight="500">${formatNum(val)}</text>\n`;
    }

    // Category label
    const cx = m.left + gi * groupWidth + groupWidth / 2;
    svg += `  <text x="${cx.toFixed(1)}" y="${m.top + ch + 22}" text-anchor="middle" fill="${T.text}" font-size="13" font-weight="500">${esc(categories[gi])}</text>\n`;
  }

  // Axis lines
  svg += `  <line x1="${m.left}" y1="${m.top}" x2="${m.left}" y2="${m.top + ch}" stroke="${T.subtle}" stroke-width="1"/>\n`;
  svg += `  <line x1="${m.left}" y1="${m.top + ch}" x2="${m.left + cw}" y2="${m.top + ch}" stroke="${T.subtle}" stroke-width="1"/>\n`;

  // Legend
  const legendY = height - 18;
  const legendTotalWidth = series.reduce((sum, s) => sum + esc(s.label).length * 7.5 + 30, 0);
  let lx = width / 2 - legendTotalWidth / 2;
  for (const s of series) {
    svg += `  <rect x="${lx.toFixed(1)}" y="${legendY - 9}" width="14" height="14" fill="${s.color}" rx="3"/>\n`;
    svg += `  <text x="${(lx + 20).toFixed(1)}" y="${legendY + 2}" fill="${T.muted}" font-size="12">${esc(s.label)}</text>\n`;
    lx += esc(s.label).length * 7.5 + 36;
  }

  svg += '</svg>\n';
  return svg;
}

// ── Summary Cards ──────────────────────────────────────────────────────
function svgSummaryCards(cards, width = 960, height = 220) {
  const cardW = (width - 40) / cards.length - 16;
  const cardH = height - 40;
  const pad = 16;

  let svg = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 ${width} ${height}" font-family="'Inter','Segoe UI',system-ui,-apple-system,sans-serif">\n`;
  svg += `  <rect width="${width}" height="${height}" fill="${T.bg}"/>\n`;

  for (let i = 0; i < cards.length; i++) {
    const c = cards[i];
    const x = 20 + i * (cardW + pad);
    const y = 20;
    const cx = x + cardW / 2;

    // Card outline
    svg += `  <rect x="${x}" y="${y}" width="${cardW}" height="${cardH}" rx="12" fill="none" stroke="${T.grid}" stroke-width="1.5"/>\n`;

    // Title
    svg += `  <text x="${cx}" y="${y + 32}" text-anchor="middle" fill="${T.muted}" font-size="13" font-weight="500">${esc(c.title)}</text>\n`;

    // Big percentage
    const isPositive = c.delta > 0;
    const arrow = isPositive ? '\u25B2' : '\u25BC';
    const color = c.goodWhenDown ? (isPositive ? '#ef4444' : '#10b981') : (isPositive ? '#10b981' : '#ef4444');
    const pct = Math.abs(c.delta).toFixed(0);
    svg += `  <text x="${cx}" y="${y + 82}" text-anchor="middle" fill="${color}" font-size="36" font-weight="800">${arrow} ${pct}%</text>\n`;

    // Before → After
    svg += `  <text x="${cx}" y="${y + 115}" text-anchor="middle" fill="${T.text}" font-size="14" font-weight="500">${esc(c.before)} \u2192 ${esc(c.after)}</text>\n`;

    // Unit
    svg += `  <text x="${cx}" y="${y + 140}" text-anchor="middle" fill="${T.muted}" font-size="11">${esc(c.unit)}</text>\n`;
  }

  svg += '</svg>\n';
  return svg;
}

// ── Journey / Flow Diagram ─────────────────────────────────────────────
function svgJourney({ title, steps, width = 800, height = 180 }) {
  const m = { top: 50, bottom: 20, side: 60 };
  const cw = width - m.side * 2;
  const boxW = 160;
  const boxH = 70;

  let svg = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 ${width} ${height}" font-family="'Inter','Segoe UI',system-ui,-apple-system,sans-serif">\n`;
  svg += `  <rect width="${width}" height="${height}" fill="${T.bg}"/>\n`;

  // Title
  svg += `  <text x="${width / 2}" y="30" text-anchor="middle" fill="${T.text}" font-size="16" font-weight="700">${esc(title)}</text>\n`;

  // Defs for arrowhead
  svg += `  <defs><marker id="arrow" viewBox="0 0 10 10" refX="9" refY="5" markerWidth="8" markerHeight="8" orient="auto-start-auto"><path d="M 0 0 L 10 5 L 0 10 z" fill="${T.muted}"/></marker></defs>\n`;

  const stepCount = steps.length;
  const spacing = cw / stepCount;

  for (let i = 0; i < stepCount; i++) {
    const s = steps[i];
    const cx = m.side + spacing * i + spacing / 2;
    const cy = m.top + boxH / 2;
    const x = cx - boxW / 2;
    const y = m.top;

    // Box
    svg += `  <rect x="${x}" y="${y}" width="${boxW}" height="${boxH}" rx="10" fill="${s.color}" opacity="0.15" stroke="${s.color}" stroke-width="1.5"/>\n`;

    // Label
    svg += `  <text x="${cx}" y="${y + 28}" text-anchor="middle" fill="${T.text}" font-size="13" font-weight="600">${esc(s.label)}</text>\n`;

    // Value
    svg += `  <text x="${cx}" y="${y + 50}" text-anchor="middle" fill="${s.color}" font-size="15" font-weight="700">${esc(s.value)}</text>\n`;

    // Arrow to next
    if (i < stepCount - 1) {
      const nextCx = m.side + spacing * (i + 1) + spacing / 2;
      const arrowY = cy;
      const ax1 = x + boxW + 4;
      const ax2 = nextCx - boxW / 2 - 4;
      const midX = (ax1 + ax2) / 2;

      svg += `  <line x1="${ax1}" y1="${arrowY}" x2="${ax2}" y2="${arrowY}" stroke="${T.muted}" stroke-width="1.5" marker-end="url(#arrow)"/>\n`;

      // Delta label above arrow
      if (s.delta) {
        svg += `  <text x="${midX}" y="${arrowY - 8}" text-anchor="middle" fill="${T.muted}" font-size="11" font-weight="500">${esc(s.delta)}</text>\n`;
      }
    }
  }

  svg += '</svg>\n';
  return svg;
}

// ── Main ───────────────────────────────────────────────────────────────
function main() {
  const data = readCsv();
  mkdirSync(OUT_DIR, { recursive: true });

  const stress = data.filter(r => r.profile === 'stress');
  const authStress = stress.filter(r => r.scenario === 'auth-flow');
  const loginStress = stress.filter(r => r.scenario === 'login-sustained');

  const load = data.filter(r => r.profile === 'load');
  const authLoad = load.filter(r => r.scenario === 'auth-flow');

  const charts = [];

  // ── 1. Throughput ────────────────────────────────────────────────────
  const throughput = svgBarChart({
    title: 'Stress Test Throughput (Total Requests)',
    categories: ['Auth-Flow', 'Login-Sustained'],
    series: [
      { label: 'baseline', color: PHASE_COLORS.baseline, values: [getVal(authStress, 'baseline', 'total_reqs'), getVal(loginStress, 'baseline', 'total_reqs')] },
      { label: 'opt-v1', color: PHASE_COLORS['opt-v1'], values: [getVal(authStress, 'opt-v1', 'total_reqs'), getVal(loginStress, 'opt-v1', 'total_reqs')] },
      { label: 'argon2id', color: PHASE_COLORS.argon2id, values: [getVal(authStress, 'argon2id', 'total_reqs'), getVal(loginStress, 'argon2id', 'total_reqs')] },
    ],
    yLabel: 'Total Requests',
  });
  writeFileSync(resolve(OUT_DIR, 'throughput-stress.svg'), throughput);
  charts.push('throughput-stress.svg');

  // ── 2. Auth-Flow Latency (Stress) ────────────────────────────────────
  const authLatency = svgBarChart({
    title: 'Auth-Flow Stress Latency',
    categories: ['baseline', 'opt-v1', 'argon2id'],
    series: [
      { label: 'p50', color: METRIC_COLORS.p50, values: [getVal(authStress, 'baseline', 'duration_p50'), getVal(authStress, 'opt-v1', 'duration_p50'), getVal(authStress, 'argon2id', 'duration_p50')] },
      { label: 'p95', color: METRIC_COLORS.p95, values: [getVal(authStress, 'baseline', 'duration_p95'), getVal(authStress, 'opt-v1', 'duration_p95'), getVal(authStress, 'argon2id', 'duration_p95')] },
      { label: 'p99', color: METRIC_COLORS.p99, values: [getVal(authStress, 'baseline', 'duration_p99'), getVal(authStress, 'opt-v1', 'duration_p99'), getVal(authStress, 'argon2id', 'duration_p99')] },
    ],
    yLabel: 'Latency (ms)',
  });
  writeFileSync(resolve(OUT_DIR, 'auth-latency-stress.svg'), authLatency);
  charts.push('auth-latency-stress.svg');

  // ── 3. Login-Sustained Latency (Stress) ──────────────────────────────
  const loginLatency = svgBarChart({
    title: 'Login-Sustained Stress Latency',
    categories: ['baseline', 'opt-v1', 'argon2id'],
    series: [
      { label: 'p50', color: METRIC_COLORS.p50, values: [getVal(loginStress, 'baseline', 'duration_p50'), getVal(loginStress, 'opt-v1', 'duration_p50'), getVal(loginStress, 'argon2id', 'duration_p50')] },
      { label: 'p95', color: METRIC_COLORS.p95, values: [getVal(loginStress, 'baseline', 'duration_p95'), getVal(loginStress, 'opt-v1', 'duration_p95'), getVal(loginStress, 'argon2id', 'duration_p95')] },
      { label: 'p99', color: METRIC_COLORS.p99, values: [getVal(loginStress, 'baseline', 'duration_p99'), getVal(loginStress, 'opt-v1', 'duration_p99'), getVal(loginStress, 'argon2id', 'duration_p99')] },
    ],
    yLabel: 'Latency (ms)',
  });
  writeFileSync(resolve(OUT_DIR, 'login-latency-stress.svg'), loginLatency);
  charts.push('login-latency-stress.svg');

  // ── 4. Average Response Time (All Scenarios, Stress) ─────────────────
  const avgResponse = svgBarChart({
    title: 'Average Response Time Under Stress',
    categories: ['Health', 'Auth-Flow', 'Login-Sustained'],
    series: [
      { label: 'baseline', color: PHASE_COLORS.baseline, values: [getVal(stress.filter(r => r.scenario === 'health'), 'baseline', 'duration_avg'), getVal(authStress, 'baseline', 'duration_avg'), getVal(loginStress, 'baseline', 'duration_avg')] },
      { label: 'opt-v1', color: PHASE_COLORS['opt-v1'], values: [getVal(stress.filter(r => r.scenario === 'health'), 'opt-v1', 'duration_avg'), getVal(authStress, 'opt-v1', 'duration_avg'), getVal(loginStress, 'opt-v1', 'duration_avg')] },
      { label: 'argon2id', color: PHASE_COLORS.argon2id, values: [0, getVal(authStress, 'argon2id', 'duration_avg'), getVal(loginStress, 'argon2id', 'duration_avg')] },
    ],
    yLabel: 'Avg Duration (ms)',
  });
  writeFileSync(resolve(OUT_DIR, 'avg-response-stress.svg'), avgResponse);
  charts.push('avg-response-stress.svg');

  // ── 5. Auth-Flow Load Latency ────────────────────────────────────────
  if (authLoad.length >= 2) {
    const authLoadLatency = svgBarChart({
      title: 'Auth-Flow Load Latency',
      categories: ['baseline', 'opt-v1'],
      series: [
        { label: 'p50', color: METRIC_COLORS.p50, values: [getVal(authLoad, 'baseline-load', 'duration_p50'), getVal(authLoad, 'opt-v1', 'duration_p50')] },
        { label: 'p95', color: METRIC_COLORS.p95, values: [getVal(authLoad, 'baseline-load', 'duration_p95'), getVal(authLoad, 'opt-v1', 'duration_p95')] },
        { label: 'p99', color: METRIC_COLORS.p99, values: [getVal(authLoad, 'baseline-load', 'duration_p99'), getVal(authLoad, 'opt-v1', 'duration_p99')] },
      ],
      yLabel: 'Latency (ms)',
    });
    writeFileSync(resolve(OUT_DIR, 'auth-latency-load.svg'), authLoadLatency);
    charts.push('auth-latency-load.svg');
  }

  // ── 6. Summary Cards ────────────────────────────────────────────────
  const summary = svgSummaryCards([
    {
      title: 'Auth p50 (stress)',
      before: '694ms', after: '95ms',
      delta: -86, goodWhenDown: true, unit: 'median latency',
    },
    {
      title: 'Login p50 (stress)',
      before: '2,607ms', after: '647ms',
      delta: -75, goodWhenDown: true, unit: 'median latency',
    },
    {
      title: 'Auth Throughput',
      before: '55K', after: '134K',
      delta: 142, goodWhenDown: false, unit: 'total requests',
    },
    {
      title: 'Login Throughput',
      before: '29K', after: '81K',
      delta: 179, goodWhenDown: false, unit: 'total requests',
    },
  ]);
  writeFileSync(resolve(OUT_DIR, 'improvements-summary.svg'), summary);
  charts.push('improvements-summary.svg');

  // ── 7. Auth-Flow Journey ─────────────────────────────────────────────
  const authJourney = svgJourney({
    title: 'Auth-Flow p50 Optimization Journey (Stress)',
    steps: [
      { label: 'Baseline', value: '694ms', color: '#ef4444', delta: '-54%' },
      { label: 'Infrastructure', value: '316ms', color: '#f59e0b', delta: '-70%' },
      { label: 'Argon2id', value: '95ms', color: '#10b981' },
    ],
  });
  writeFileSync(resolve(OUT_DIR, 'auth-journey.svg'), authJourney);
  charts.push('auth-journey.svg');

  // ── 8. Login Journey ─────────────────────────────────────────────────
  const loginJourney = svgJourney({
    title: 'Login-Sustained p50 Optimization Journey (Stress)',
    steps: [
      { label: 'Baseline', value: '2,607ms', color: '#ef4444', delta: '-2%' },
      { label: 'Infrastructure', value: '2,553ms', color: '#f59e0b', delta: '-75%' },
      { label: 'Argon2id', value: '647ms', color: '#10b981' },
    ],
  });
  writeFileSync(resolve(OUT_DIR, 'login-journey.svg'), loginJourney);
  charts.push('login-journey.svg');

  // ── Done ─────────────────────────────────────────────────────────────
  console.log(`\n  Generated ${charts.length} charts (${theme} theme) in benchmark/graphs/\n`);
  for (const c of charts) console.log(`    ${c}`);
  console.log();
}

main();
