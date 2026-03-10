#!/usr/bin/env node
import { execFileSync, spawn } from 'node:child_process';
import { createReadStream, readFileSync, appendFileSync, existsSync, mkdirSync, readdirSync } from 'node:fs';
import { createInterface } from 'node:readline';
import { basename, join, resolve } from 'node:path';

// ── Constants ──────────────────────────────────────────────────────────
const SCENARIOS = ['health', 'auth-flow', 'login-sustained'];
const PROFILES = ['smoke', 'load', 'stress', 'spike', 'breakpoint'];
const RESULTS_DIR = resolve('benchmark/results');
const CSV_PATH = resolve('benchmark/results.csv');
const CSV_HEADER = 'timestamp,scenario,profile,label,total_reqs,rps,duration_avg,duration_p50,duration_p95,duration_p99,errors';
const COOLDOWN_SECS = 10;

// ── Stats ──────────────────────────────────────────────────────────────
function percentile(values, p) {
  if (values.length === 0) return 0;
  const sorted = Float64Array.from(values).sort();
  const idx = Math.max(0, Math.ceil((p / 100) * sorted.length) - 1);
  return sorted[idx];
}

function avg(values) {
  if (values.length === 0) return 0;
  let sum = 0;
  for (const v of values) sum += v;
  return sum / values.length;
}

// ── k6 JSON Parser ────────────────────────────────────────────────────
async function parseK6File(filePath) {
  const metrics = {};
  const rl = createInterface({ input: createReadStream(filePath), crlfDelay: Infinity });

  for await (const line of rl) {
    try {
      const obj = JSON.parse(line);
      if (obj.type === 'Point' && obj.metric) {
        (metrics[obj.metric] ??= []).push(obj.data.value);
      }
    } catch {}
  }

  const durations = metrics['http_req_duration'] ?? [];
  const reqs = metrics['http_reqs'] ?? [];
  const iters = metrics['iteration_duration'] ?? [];
  const errors = metrics['errors'] ?? [];

  let rps = 0;
  if (iters.length > 0) {
    const avgIterMs = avg(iters);
    if (avgIterMs > 0) rps = reqs.length / (iters.length * avgIterMs / 1000);
  }

  const { scenario, profile, ts } = parseFilename(filePath);

  return {
    timestamp: ts,
    scenario,
    profile,
    total_reqs: reqs.length,
    rps: +rps.toFixed(2),
    duration_avg: +avg(durations).toFixed(2),
    duration_p50: +percentile(durations, 50).toFixed(2),
    duration_p95: +percentile(durations, 95).toFixed(2),
    duration_p99: +percentile(durations, 99).toFixed(2),
    errors: errors.length,
  };
}

function parseFilename(filePath) {
  const base = basename(filePath, '.json');
  const parts = base.split('_');
  const profileIdx = parts.findIndex((p) => PROFILES.includes(p));

  if (profileIdx === -1) return { scenario: base, profile: 'unknown', ts: timestamp() };

  const scenario = parts.slice(0, profileIdx).join('-');
  const profile = parts[profileIdx];
  const rest = parts.slice(profileIdx + 1);
  const ts = rest.length >= 2 ? `${rest[0]}_${rest[1]}` : rest[0] ?? timestamp();

  return { scenario, profile, ts };
}

// ── CSV Operations ─────────────────────────────────────────────────────
function appendCsv(result, label) {
  const needsHeader = !existsSync(CSV_PATH);
  if (needsHeader) appendFileSync(CSV_PATH, CSV_HEADER + '\n');

  const row = [
    result.timestamp, result.scenario, result.profile, label,
    result.total_reqs, result.rps,
    result.duration_avg, result.duration_p50, result.duration_p95, result.duration_p99,
    result.errors,
  ].join(',');

  appendFileSync(CSV_PATH, row + '\n');
}

function readCsv() {
  if (!existsSync(CSV_PATH)) return [];
  const lines = readFileSync(CSV_PATH, 'utf8').trim().split('\n').slice(1);
  return lines.map((line) => {
    const [ts, scenario, profile, label, total_reqs, rps, avg, p50, p95, p99, errors] = line.split(',');
    return { timestamp: ts, scenario, profile, label, total_reqs: +total_reqs, rps: +rps, duration_avg: +avg, duration_p50: +p50, duration_p95: +p95, duration_p99: +p99, errors: +errors };
  });
}

function findPrevious(scenario, profile, currentTimestamp) {
  const rows = readCsv().filter((r) => r.scenario === scenario && r.profile === profile && r.timestamp !== currentTimestamp);
  return rows.length > 0 ? rows[rows.length - 1] : null;
}

// ── Comparison Display ─────────────────────────────────────────────────
function printResult(result, label) {
  const header = `${result.scenario} × ${result.profile}` + (label ? ` [${label}]` : '');
  console.log(`\n  ${header}`);
  console.log('  ' + '─'.repeat(50));
  console.log(`  ${'Reqs'.padEnd(12)} ${String(result.total_reqs).padStart(10)}`);
  console.log(`  ${'RPS'.padEnd(12)} ${result.rps.toFixed(2).padStart(10)}`);
  console.log(`  ${'Avg'.padEnd(12)} ${(result.duration_avg.toFixed(2) + 'ms').padStart(10)}`);
  console.log(`  ${'p50'.padEnd(12)} ${(result.duration_p50.toFixed(2) + 'ms').padStart(10)}`);
  console.log(`  ${'p95'.padEnd(12)} ${(result.duration_p95.toFixed(2) + 'ms').padStart(10)}`);
  console.log(`  ${'p99'.padEnd(12)} ${(result.duration_p99.toFixed(2) + 'ms').padStart(10)}`);
  console.log(`  ${'Errors'.padEnd(12)} ${String(result.errors).padStart(10)}`);
}

function printComparison(current, previous, label) {
  const header = `${current.scenario} × ${current.profile}` + (label ? ` [${label}]` : '');
  console.log(`\n  ${header}`);
  console.log('  ' + '─'.repeat(60));

  const hdr = `  ${'Metric'.padEnd(10)} ${'Previous'.padStart(12)} ${'Current'.padStart(12)} ${'Delta'.padStart(20)}`;
  console.log(hdr);
  console.log('  ' + '─'.repeat(60));

  const rows = [
    ['RPS', previous.rps, current.rps, true],
    ['Avg', previous.duration_avg, current.duration_avg, false],
    ['p50', previous.duration_p50, current.duration_p50, false],
    ['p95', previous.duration_p95, current.duration_p95, false],
    ['p99', previous.duration_p99, current.duration_p99, false],
    ['Errors', previous.errors, current.errors, false],
  ];

  for (const [name, prev, curr, higherIsBetter] of rows) {
    const diff = curr - prev;
    const pct = prev !== 0 ? ((diff / prev) * 100).toFixed(1) : '—';
    const sign = diff > 0 ? '+' : '';
    const unit = name === 'RPS' || name === 'Errors' ? '' : 'ms';
    const delta = prev === 0 && curr === 0 ? '—' : `${sign}${diff.toFixed(2)} (${pct === '—' ? '—' : sign + pct + '%'})`;
    console.log(`  ${name.padEnd(10)} ${(prev.toFixed(2) + unit).padStart(12)} ${(curr.toFixed(2) + unit).padStart(12)} ${delta.padStart(20)}`);
  }
  console.log('  ' + '─'.repeat(60));
}

// ── Commands ───────────────────────────────────────────────────────────
async function cmdRun(args) {
  const scenario = args[0];
  const profile = args[1] ?? 'load';
  const label = parseFlag(args, '--label') ?? parseFlag(args, '-l') ?? '';

  if (!scenario || !SCENARIOS.includes(scenario)) {
    console.error(`Invalid scenario: ${scenario}\nAvailable: ${SCENARIOS.join(', ')}`);
    process.exit(1);
  }
  if (!PROFILES.includes(profile)) {
    console.error(`Invalid profile: ${profile}\nAvailable: ${PROFILES.join(', ')}`);
    process.exit(1);
  }

  const k6 = findK6();
  mkdirSync(RESULTS_DIR, { recursive: true });

  const ts = timestamp();
  const outputFile = join(RESULTS_DIR, `${scenario}_${profile}_${ts}.json`);
  const scenarioFile = resolve(`benchmark/scenarios/${scenario}.js`);
  const baseUrl = process.env.BASE_URL ?? 'http://localhost:8080';

  console.log(`\n  === Benchmark: ${scenario} | Profile: ${profile} ===`);
  console.log(`      Target: ${baseUrl}`);
  console.log(`      Output: ${outputFile}\n`);

  const code = await new Promise((res) => {
    const child = spawn(k6, [
      'run',
      '-e', `PROFILE=${profile}`,
      '-e', `BASE_URL=${baseUrl}`,
      '--out', `json=${outputFile}`,
      scenarioFile,
    ], { stdio: 'inherit' });
    child.on('close', res);
  });

  if (code !== 0) {
    console.error(`\n  k6 exited with code ${code}`);
    process.exit(code);
  }

  console.log('\n  Extracting results...');
  const result = await parseK6File(outputFile);
  appendCsv(result, label);

  const previous = findPrevious(result.scenario, result.profile, result.timestamp);
  if (previous) {
    printComparison(result, previous, label);
  } else {
    printResult(result, label);
    console.log('\n  (No previous run to compare against)');
  }

  console.log(`\n  Appended to ${CSV_PATH}`);
}

async function cmdRunAll(args) {
  const profile = args[0] ?? 'load';
  const label = parseFlag(args, '--label') ?? parseFlag(args, '-l') ?? '';

  console.log(`\n  === Running all benchmarks: ${profile} ===\n`);

  for (let i = 0; i < SCENARIOS.length; i++) {
    await cmdRun([SCENARIOS[i], profile, ...(label ? ['--label', label] : [])]);

    if (i < SCENARIOS.length - 1) {
      console.log(`\n  --- Cooldown (${COOLDOWN_SECS}s) ---`);
      await new Promise((r) => setTimeout(r, COOLDOWN_SECS * 1000));
    }
  }

  console.log('\n  === All benchmarks complete ===');
}

async function cmdExtract(args) {
  const pattern = args[0];
  const label = parseFlag(args, '--label') ?? parseFlag(args, '-l') ?? '';

  if (!pattern) {
    console.error('Usage: node bench.mjs extract <glob> [--label <tag>]');
    process.exit(1);
  }

  const lastSlash = Math.max(pattern.lastIndexOf('/'), pattern.lastIndexOf('\\'));
  const dir = resolve(lastSlash >= 0 ? pattern.substring(0, lastSlash) : '.');
  const filePattern = lastSlash >= 0 ? pattern.substring(lastSlash + 1) : pattern;
  const regex = new RegExp('^' + filePattern.replace(/\./g, '\\.').replace(/\*/g, '.*').replace(/\?/g, '.') + '$');

  let files;
  try {
    files = readdirSync(dir)
      .filter((f) => regex.test(f))
      .sort()
      .map((f) => join(dir, f));
  } catch {
    console.error(`Cannot read directory: ${dir}`);
    process.exit(1);
  }

  if (files.length === 0) {
    console.error(`No files matched: ${pattern}`);
    process.exit(1);
  }

  console.log(`  Matched ${files.length} file(s)\n`);

  let errors = 0;
  for (const file of files) {
    try {
      const result = await parseK6File(file);
      appendCsv(result, label);
      printResult(result, label);
    } catch (err) {
      console.error(`  SKIP ${file}: ${err.message}`);
      errors++;
    }
  }

  if (errors > 0) {
    console.error(`\n  ${errors} file(s) failed`);
    process.exit(1);
  }

  console.log(`\n  Appended to ${CSV_PATH}`);
}

async function cmdCompare(args) {
  const scenario = args[0];
  const profile = args[1];

  if (!scenario || !profile) {
    console.error('Usage: node bench.mjs compare <scenario> <profile>');
    process.exit(1);
  }

  const rows = readCsv().filter((r) => r.scenario === scenario && r.profile === profile);

  if (rows.length < 2) {
    console.error(`Need at least 2 runs of ${scenario}/${profile} to compare. Found ${rows.length}.`);
    process.exit(1);
  }

  const current = rows[rows.length - 1];
  const previous = rows[rows.length - 2];
  printComparison(current, previous, current.label);
}

// ── CLI Entry ──────────────────────────────────────────────────────────
const [cmd, ...args] = process.argv.slice(2);

const commands = { run: cmdRun, 'run-all': cmdRunAll, extract: cmdExtract, compare: cmdCompare };

if (!cmd || !commands[cmd]) {
  console.error(`Usage: node bench.mjs <command>

Commands:
  run <scenario> <profile> [--label <tag>]     Run k6 → extract → compare
  run-all <profile> [--label <tag>]            Run all scenarios sequentially
  extract <glob> [--label <tag>]               Extract existing JSON files to CSV
  compare <scenario> <profile>                 Compare last two runs from CSV

Scenarios: ${SCENARIOS.join(', ')}
Profiles:  ${PROFILES.join(', ')}`);
  process.exit(1);
}

// ── k6 Binary Discovery ───────────────────────────────────────────────
function findK6() {
  try {
    execFileSync('k6', ['version'], { stdio: 'ignore' });
    return 'k6';
  } catch {}

  const winPath = 'C:\\Program Files\\k6\\k6.exe';
  try {
    execFileSync(winPath, ['version'], { stdio: 'ignore' });
    return winPath;
  } catch {}

  console.error('Error: k6 not found. Install with: winget install k6');
  process.exit(1);
}

// ── Arg Helpers ────────────────────────────────────────────────────────
function parseFlag(args, flag) {
  const i = args.indexOf(flag);
  if (i === -1) return undefined;
  return args[i + 1];
}

function timestamp() {
  const d = new Date();
  const pad = (n) => String(n).padStart(2, '0');
  return `${d.getFullYear()}${pad(d.getMonth() + 1)}${pad(d.getDate())}_${pad(d.getHours())}${pad(d.getMinutes())}${pad(d.getSeconds())}`;
}

commands[cmd](args);
