#!/usr/bin/env bun
// Guard that codecov.yml's `after_n_builds` equals the number of coverage
// uploads ci.yml actually performs per commit. See issue #4991.
//
// codecov の after_n_builds は既定が 1 で、**最初のアップロードが届いた時点で
// ステータスを確定させる**。フロントを 4 シャードに分けている本リポジトリでは、
// 最初に到着したシャードが担当していないファイルが 0% のまま採点されていた。
// PR #5068 の TressettePage.tsx と PR #5073 の SlapjackPage.tsx はどちらも 0% と
// 報告され、ローカル計測では該当行に 124 hits あり、全シャード到着後には同じ
// ステータスが自然に green へ変わっていた。
//
// 直し方は「全部そろうまで待たせる」だけだが、待たせる本数はジョブ構成に
// 依存する。シャードを 4 から 8 に増やしたり、アップロードを 1 つ足したりすると
// codecov.yml の数字は黙って古くなる。**数字が小さすぎれば偽陽性が戻り、
// 大きすぎればステータスが永久に pending になる** — どちらも CI は緑のままで、
// 気づくのは次に誰かが 0% を踏んだときになる。
//
// そこで ci.yml を実際に読み、
//   - codecov/codecov-action を使うステップ数
//   - そのステップを含むジョブの matrix シャード数
// から本数を組み立てて突き合わせる。

import { readFile } from 'node:fs/promises';
import { join } from 'node:path';
import { fileURLToPath } from 'node:url';
import { assertFloor } from './lib/floor.mjs';

const FRONTEND = fileURLToPath(new URL('..', import.meta.url));
const REPO = join(FRONTEND, '..');
// テストからフィクスチャを差し込めるように環境変数で上書き可能にする。ガードは
// 「正しい入力で鳴らない」ことも確かめないと、全部落とす実装と見分けが付かない。
const CI = process.env.CHECK_CODECOV_CI ?? join(REPO, '.github/workflows/ci.yml');
const CODECOV = process.env.CHECK_CODECOV_YML ?? join(REPO, 'codecov.yml');

/**
 * ci.yml をジョブ単位に切り出す。`jobs:` 直下の 2 スペースインデントのキーが
 * ジョブ名で、次のジョブ名までがその本文。
 *
 * @param {string} src - ci.yml の中身。
 * @returns {{name: string, body: string}[]} ジョブ名と本文の配列。
 */
function splitJobs(src) {
  const lines = src.split('\n');
  const start = lines.findIndex((l) => l === 'jobs:');
  if (start < 0) {
    throw new Error('check-codecov-uploads: ci.yml に `jobs:` が無い');
  }
  const jobs = [];
  let current = null;
  for (const line of lines.slice(start + 1)) {
    const header = /^ {2}([A-Za-z0-9_-]+):\s*$/.exec(line);
    if (header) {
      current = { name: header[1], body: '' };
      jobs.push(current);
      continue;
    }
    if (current) current.body += `${line}\n`;
  }
  return jobs;
}

/**
 * ジョブ本文から matrix のシャード数を読む。`shard: [1, 2, 3, 4]` を想定し、
 * matrix が無ければ 1 本。
 *
 * @param {string} body - ジョブ本文。
 * @returns {number} そのジョブが展開される本数。
 */
function shardCount(body) {
  const m = /^\s*shard:\s*\[([^\]]*)\]\s*$/m.exec(body);
  if (!m) return 1;
  const n = m[1].split(',').filter((s) => s.trim() !== '').length;
  if (n < 1) {
    throw new Error('check-codecov-uploads: shard 配列が空');
  }
  return n;
}

const ci = await readFile(CI, 'utf8');
const jobs = splitJobs(ci);
// フィクスチャはジョブ数が少ないので、床は本物の ci.yml を読んだときだけ。
if (!process.env.CHECK_CODECOV_CI) {
  assertFloor('check-codecov-uploads', jobs.length, 5, 'jobs parsed from ci.yml');
}

const uploaders = [];
for (const job of jobs) {
  const uses = job.body.split('\n').filter((l) => l.includes('codecov/codecov-action')).length;
  if (uses === 0) continue;
  uploaders.push({ job: job.name, uses, shards: shardCount(job.body) });
}
assertFloor('check-codecov-uploads', uploaders.length, 1, 'jobs uploading to Codecov');

const expected = uploaders.reduce((sum, u) => sum + u.uses * u.shards, 0);

const codecov = await readFile(CODECOV, 'utf8');
// `codecov: > notify: > after_n_builds` (docs.codecov.com/docs/codecovyml-reference)
const declared = /^\s*after_n_builds:\s*(\d+)\s*$/m.exec(codecov);
if (!declared) {
  console.error(
    'check-codecov-uploads: codecov.yml に after_n_builds がありません。\n' +
      `  ci.yml のアップロードは ${expected} 本です。既定の 1 のままだと、最初に\n` +
      '  届いたシャードだけでステータスが確定し、そのシャードが担当していない\n' +
      '  ファイルが 0% と誤報されます (#4991)。',
  );
  process.exit(1);
}

const declaredN = Number(declared[1]);
if (declaredN !== expected) {
  const breakdown = uploaders.map((u) => `    ${u.job}: ${u.uses} x ${u.shards} shard(s)`).join('\n');
  console.error(
    `check-codecov-uploads: after_n_builds=${declaredN} ですが、ci.yml のアップロードは ${expected} 本です。\n` +
      `${breakdown}\n` +
      '  小さすぎると 0% の偽陽性が戻り、大きすぎるとステータスが永久に pending に\n' +
      '  なります。codecov.yml の数字を合わせてください (#4991)。',
  );
  process.exit(1);
}

console.log(
  `check-codecov-uploads: OK (after_n_builds=${declaredN} == ${expected} uploads across ${uploaders.length} job(s)).`,
);
