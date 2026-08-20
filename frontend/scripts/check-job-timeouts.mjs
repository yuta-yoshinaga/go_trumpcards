#!/usr/bin/env bun
// Guard that every ci.yml job declares a timeout-minutes. See issue #5967.
//
// GitHub の既定は **6 時間**。ハングしたジョブはその間ランナーを掴んだまま、
// PR は「pending 1」でマージゲートを通れない。人が気づいて手でキャンセルする
// までは止まらない (#5967 は E2E が 26 / 27 / 46 分走り続けた 3 例)。
//
// 上限そのものは ci.yml のコメントに実測値と一緒に書いてある。ここが見るのは
// **「新しく足したジョブに上限が付いているか」**だけ ── 数字の妥当性は人が
// 決めるが、付け忘れは機械で防げる。

import { readFile } from 'node:fs/promises';
import { join } from 'node:path';
import { fileURLToPath } from 'node:url';
import { assertFloor } from './lib/floor.mjs';

const FRONTEND = fileURLToPath(new URL('..', import.meta.url));
const REPO = join(FRONTEND, '..');
// テストからフィクスチャを差し込めるようにする。ガードは「正しい入力で鳴らない」
// ことも確かめないと、全部落とす実装と見分けが付かない。
const CI = process.env.CHECK_JOB_TIMEOUTS_CI ?? join(REPO, '.github/workflows/ci.yml');

/**
 * ci.yml をジョブ単位に切り出す。`jobs:` 直下の 2 スペースインデントのキーが
 * ジョブ名で、次のジョブ名までがその本文。
 *
 * @param {string} src - ci.yml の中身。
 * @returns {{name: string, body: string}[]} ジョブ名と本文の配列。
 */
export function splitJobs(src) {
  const lines = src.split('\n');
  const start = lines.indexOf('jobs:');
  if (start < 0) {
    throw new Error('check-job-timeouts: ci.yml に `jobs:` が無い');
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
 * 上限の付いていないジョブ名を返す。
 *
 * @param {string} src - ci.yml の中身。
 * @returns {string[]} `timeout-minutes` を持たないジョブ名。
 */
export function jobsWithoutTimeout(src) {
  return (
    splitJobs(src)
      // **ジョブ直下の 4 スペース固定で見る。**任意のインデントを許すと、ステップに
      // 付いた `timeout-minutes` でジョブ全体が有界だと誤判定する。ステップ側の上限は
      // そのステップしか止めないので、別のステップで詰まればジョブは 6 時間残る。
      .filter((job) => !/^ {4}timeout-minutes: *\d+ *$/m.test(job.body))
      .map((job) => job.name)
  );
}

/**
 * OS パッケージを入れるステップのうち、ステップ単位の上限が無いものを返す。
 *
 * apt はロック待ちや対話プロンプトで固まる。ジョブ側の上限しか無いと、
 * **テストが 1 本も走らないままジョブの予算を丸ごと使い切って cancel** される
 * (#6002 で 2 本の PR が連続で踏んだ)。ステップ側で先に切り上げれば、残りの
 * 予算はテストに残る。
 *
 * @param {string} src - ci.yml の中身。
 * @returns {string[]} 上限の無い apt ステップ名 (無名なら run 行)。
 */
export function aptStepsWithoutTimeout(src) {
  const steps = src.split(/^ {6}- /m).slice(1);
  return (
    steps
      // **コメント行は見ない。**次のステップに付けた説明コメントは前のステップの
      // 塊に入るので、素の全文検索だと「apt と書いてある解説」で隣のステップを
      // 誤検知する (実際に一度誤検知した)。
      .filter((step) =>
        step
          .split('\n')
          .filter((line) => !/^\s*#/.test(line))
          .some((line) => /(?:apt-get|apt |install-deps|--with-deps)/.test(line)),
      )
      .filter((step) => !/^ {8}timeout-minutes: *\d+ *$/m.test(step))
      .map((step) => {
        const named = /^name: (.+)$/m.exec(step);
        const run = /^ *run: (.+)$/m.exec(step);
        return named?.[1] ?? run?.[1] ?? '(unnamed step)';
      })
  );
}

const src = await readFile(CI, 'utf8');
const jobs = splitJobs(src);
// **0 件で成功と読ませない。**`jobs:` の切り出しが壊れれば「違反 0 件」になり、
// 何も見ていない状態が緑で通る。本番は 7 ジョブなので、床は普段の増減では
// 踏まず、走査が壊れたときだけ踏む 5 に置く。
assertFloor('check-job-timeouts', jobs.length, 5, 'jobs');
const missing = jobsWithoutTimeout(src);
if (missing.length > 0) {
  console.error(`check-job-timeouts: timeout-minutes の無いジョブ: ${missing.join(', ')}`);
  console.error('  ハングしても既定の 6 時間止まりません。実測 max のおよそ 2 倍を目安に付けてください。');
  process.exit(1);
}
const unboundedApt = aptStepsWithoutTimeout(src);
if (unboundedApt.length > 0) {
  console.error(`check-job-timeouts: ステップ単位の timeout-minutes が無い apt ステップ: ${unboundedApt.join(', ')}`);
  console.error('  apt が固まるとジョブの予算を丸ごと使い切り、テストが 1 本も走らずに cancel されます (#6002)。');
  process.exit(1);
}
console.log(`job-timeouts: OK (${jobs.length.toString()} jobs bounded, apt steps bounded).`);
