import { spawnSync } from 'node:child_process';
import { existsSync, mkdtempSync, rmSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { afterAll, describe, expect, it } from 'vitest';

// vitest の root は `frontend/` なので cwd から解決する。パスを間違えると全ケースが
// 「無いファイルを spawn した」になり、落ちてほしいケースが落ちて見えるだけになる。
const SCRIPTS = join(process.cwd(), 'scripts');
const GUARD = join(SCRIPTS, 'check-job-timeouts.mjs');
if (!existsSync(GUARD)) throw new Error(`check-job-timeouts.mjs not found at ${GUARD} (cwd: ${process.cwd()})`);

const dirs = [];
afterAll(() => {
  for (const d of dirs) rmSync(d, { recursive: true, force: true });
});

/** 上限の付いた 5 ジョブ。ジョブ数の下限 (5) をちょうど満たす形。 */
const CI_BOUNDED = `name: CI
jobs:
  test-hooks:
    runs-on: ubuntu-latest
    timeout-minutes: 10
    steps:
      - run: echo hi
  lint-backend:
    runs-on: ubuntu-latest
    timeout-minutes: 15
    steps:
      - run: echo hi
  test-backend:
    runs-on: ubuntu-latest
    timeout-minutes: 20
    steps:
      - run: echo hi
  test-frontend:
    runs-on: ubuntu-latest
    timeout-minutes: 20
    strategy:
      matrix:
        shard: [1, 2]
    steps:
      - run: echo hi
  test-e2e:
    runs-on: ubuntu-latest
    timeout-minutes: 20
    steps:
      - run: echo hi
`;

function runGuard(ci) {
  const dir = mkdtempSync(join(tmpdir(), 'job-timeouts-'));
  dirs.push(dir);
  const path = join(dir, 'ci.yml');
  writeFileSync(path, ci);
  return spawnSync('bun', [GUARD], { env: { ...process.env, CHECK_JOB_TIMEOUTS_CI: path }, encoding: 'utf8' });
}

describe('check-job-timeouts', () => {
  it('passes when every job is bounded', () => {
    const r = runGuard(CI_BOUNDED);
    expect(r.status).toBe(0);
    expect(r.stdout).toContain('5 jobs');
  });

  it('fails and names the job that has no timeout', () => {
    const r = runGuard(
      CI_BOUNDED.replace(
        '  test-e2e:\n    runs-on: ubuntu-latest\n    timeout-minutes: 20\n',
        '  test-e2e:\n    runs-on: ubuntu-latest\n',
      ),
    );
    expect(r.status).toBe(1);
    expect(r.stderr).toContain('test-e2e');
    // 他の 4 本は巻き込まない ── 名指しが効いていることの確認。
    expect(r.stderr).not.toContain('test-backend');
  });

  // レビュー #5969: ステップに付いた上限はそのステップしか止めない。ジョブ全体は
  // 既定の 6 時間のままなので、これを「有界」と読んではいけない。
  it('does not accept a step-level timeout as a job-level bound', () => {
    const stepOnly = CI_BOUNDED.replace(
      '  test-e2e:\n    runs-on: ubuntu-latest\n    timeout-minutes: 20\n    steps:\n      - run: echo hi\n',
      '  test-e2e:\n    runs-on: ubuntu-latest\n    steps:\n      - run: echo hi\n        timeout-minutes: 20\n',
    );
    const r = runGuard(stepOnly);
    expect(r.status).toBe(1);
    expect(r.stderr).toContain('test-e2e');
  });

  // **0 件で成功と読ませない側。**切り出しが壊れた ci.yml は「違反 0 件」に
  // なるので、ジョブ数の下限で落ちること自体を確かめる。
  it('fails when it cannot find enough jobs to have checked anything', () => {
    const r = runGuard(
      'name: CI\njobs:\n  only-one:\n    runs-on: ubuntu-latest\n    timeout-minutes: 5\n    steps:\n      - run: echo hi\n',
    );
    expect(r.status).toBe(1);
    expect(r.stderr).toContain('only 1 jobs found');
  });

  it('fails when there is no jobs block at all', () => {
    const r = runGuard('name: CI\non: push\n');
    expect(r.status).not.toBe(0);
  });

  // 本番の ci.yml でも通ること。フィクスチャだけ緑では意味がない。
  it('passes against the real ci.yml', () => {
    const r = spawnSync('bun', [GUARD], { encoding: 'utf8' });
    expect(r.status).toBe(0);
  });
});
