import { spawnSync } from 'node:child_process';
import { existsSync, mkdtempSync, rmSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { afterAll, describe, expect, it } from 'vitest';

// vitest の root は `frontend/` なので cwd から解決する。パスを間違えると全ケースが
// 「無いファイルを spawn した」になり、落ちてほしいケースが落ちて見えるだけになる。
const SCRIPTS = join(process.cwd(), 'scripts');
const GUARD = join(SCRIPTS, 'check-codecov-uploads.mjs');
if (!existsSync(GUARD)) throw new Error(`check-codecov-uploads.mjs not found at ${GUARD} (cwd: ${process.cwd()})`);

const dirs = [];
afterAll(() => {
  for (const d of dirs) rmSync(d, { recursive: true, force: true });
});

/** backend 1 本 + frontend 4 シャード = 5 本、という本番と同じ形。 */
const CI_FIVE = `name: CI
jobs:
  test-backend:
    steps:
      - name: Upload backend coverage to Codecov
        uses: codecov/codecov-action@v6
  test-frontend:
    strategy:
      matrix:
        shard: [1, 2, 3, 4]
    steps:
      - name: Upload frontend coverage to Codecov
        uses: codecov/codecov-action@v6
  test-e2e:
    strategy:
      matrix:
        shard: [1, 2, 3]
    steps:
      - name: Run Playwright
        run: bunx playwright test
`;

/**
 * フィクスチャを書いてガードを走らせる。
 *
 * @param {string} ci - ci.yml の中身。
 * @param {string} codecov - codecov.yml の中身。
 * @returns {{status: number, out: string}} 終了コードと stdout+stderr。
 */
function run(ci, codecov) {
  const dir = mkdtempSync(join(tmpdir(), 'codecov-guard-'));
  dirs.push(dir);
  const ciPath = join(dir, 'ci.yml');
  const ccPath = join(dir, 'codecov.yml');
  writeFileSync(ciPath, ci);
  writeFileSync(ccPath, codecov);
  const r = spawnSync('bun', [GUARD], {
    encoding: 'utf8',
    env: { ...process.env, CHECK_CODECOV_CI: ciPath, CHECK_CODECOV_YML: ccPath },
  });
  return { status: r.status, out: `${r.stdout}${r.stderr}` };
}

describe('check-codecov-uploads', () => {
  // **正しい入力で鳴らないことを先に確かめる。**これが無いと、何を渡しても
  // 落ちる実装と見分けが付かない。
  it('passes when after_n_builds matches the upload count', () => {
    const { status, out } = run(CI_FIVE, 'codecov:\n  notify:\n    after_n_builds: 5\n');
    expect(status).toBe(0);
    expect(out).toContain('OK');
  });

  it('fails when after_n_builds is too small (the 0% false positive returns)', () => {
    const { status, out } = run(CI_FIVE, 'codecov:\n  notify:\n    after_n_builds: 1\n');
    expect(status).toBe(1);
    expect(out).toContain('5 本');
  });

  it('fails when after_n_builds is too large (the status would hang pending)', () => {
    const { status } = run(CI_FIVE, 'codecov:\n  notify:\n    after_n_builds: 9\n');
    expect(status).toBe(1);
  });

  it('fails when after_n_builds is missing entirely', () => {
    const { status, out } = run(CI_FIVE, 'coverage:\n  status:\n    patch:\n      default:\n        target: 80%\n');
    expect(status).toBe(1);
    expect(out).toContain('after_n_builds');
  });

  // **これが本命。**数字が古くなる現実的な経路はシャード数の変更であって、
  // codecov.yml を直接いじることではない。
  it('fails when ci.yml changes its shard count and codecov.yml is left behind', () => {
    const eightShards = CI_FIVE.replace('shard: [1, 2, 3, 4]', 'shard: [1, 2, 3, 4, 5, 6, 7, 8]');
    const { status, out } = run(eightShards, 'codecov:\n  notify:\n    after_n_builds: 5\n');
    expect(status).toBe(1);
    expect(out).toContain('9 本');
  });

  // E2E のように codecov へ投げないシャードジョブを数えてしまうと、本数が
  // 過大になってステータスが永久に pending になる。
  it('ignores sharded jobs that do not upload to Codecov', () => {
    const { status, out } = run(CI_FIVE, 'codecov:\n  notify:\n    after_n_builds: 5\n');
    expect(status).toBe(0);
    expect(out).not.toContain('test-e2e');
  });

  // **コメントを数えない。**ci.yml には長い日本語コメントを書く流儀があるので、
  // アクション名に触れた説明が1行増えただけで本数が水増しされると、
  // ステータスが永久に pending になる。
  it('counts only `uses:` lines, not comments mentioning the action', () => {
    const withComment = CI_FIVE.replace(
      '      - name: Upload backend coverage to Codecov',
      '      # codecov/codecov-action へのアップロードはここだけ\n      - name: Upload backend coverage to Codecov',
    );
    const { status, out } = run(withComment, 'codecov:\n  notify:\n    after_n_builds: 5\n');
    expect(status).toBe(0);
    expect(out).toContain('OK');
  });
});
