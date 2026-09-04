import { spawnSync } from 'node:child_process';
import { existsSync, mkdirSync, mkdtempSync, rmSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { afterAll, describe, expect, it } from 'vitest';

const SCRIPTS = join(process.cwd(), 'scripts');
const GUARD = join(SCRIPTS, 'check-conditional-testid-coverage.mjs');
if (!existsSync(GUARD)) {
  throw new Error(`guard not found at ${GUARD} (cwd: ${process.cwd()})`);
}

const dirs = [];
afterAll(() => {
  for (const d of dirs) rmSync(d, { recursive: true, force: true });
});

/**
 * Builds a fixture directory and runs check-conditional-testid-coverage.mjs against it.
 *
 * @param {Record<string, string>} files - Map of relative paths to file contents.
 * @returns {{ status: number | null, stdout: string, stderr: string }}
 */
function run(files) {
  const root = mkdtempSync(join(tmpdir(), 'cond-testid-guard-'));
  dirs.push(root);
  for (const [relPath, content] of Object.entries(files)) {
    const full = join(root, relPath);
    mkdirSync(join(full, '..'), { recursive: true });
    writeFileSync(full, content);
  }
  const r = spawnSync(process.execPath, [GUARD, '--root', root], {
    encoding: 'utf8',
    cwd: process.cwd(),
  });
  return { status: r.status, stdout: r.stdout, stderr: r.stderr };
}

describe('check-conditional-testid-coverage', () => {
  // ページだけを見る guard は、共有コンポーネント側の同じ欠陥を通す。
  // 面を 1 つしか見ない guard は、もう 1 つの面で必ず破られる。
  it('rejects an unreferenced conditional testid in src/components as well', () => {
    const r = run({
      'src/components/SampleCard.tsx': `
        export function SampleCard({ showBadge }) {
          return (
            <div>
              {showBadge && (
                <span data-testid="unreferenced-component-badge">Badge</span>
              )}
            </div>
          );
        }
      `,
      'src/components/SampleCard.test.tsx': `
        it('renders card', () => {
          render(<SampleCard showBadge={false} />);
        });
      `,
    });

    expect(r.status).toBe(1);
    expect(r.stderr).toContain('SampleCard.tsx');
    expect(r.stderr).toContain('unreferenced-component-badge');
  });

  // フィクスチャ用の root は `--root` でしか渡せない。裸の位置引数でも root に
  // なってしまうと、**引数を 1 つ足すだけで floor が黙って無効になる**。
  // floor は「走査が壊れていないこと」を担保する唯一の仕掛けなので、そこは塞ぐ。
  it('ignores a bare positional path and still scans the repo (floors stay armed)', () => {
    const root = mkdtempSync(join(tmpdir(), 'cond-testid-stray-'));
    dirs.push(root);
    const r = spawnSync(process.execPath, [GUARD, root], { encoding: 'utf8', cwd: process.cwd() });
    // 空ディレクトリを root に採ったなら "0 page components" になる。実リポジトリを
    // 走査していれば数百ページになる。後者であることを見る。
    expect(r.stdout).not.toContain('0 page/component sources scanned');
    expect(r.stdout).toMatch(/[1-9]\d{2,} page\/component sources scanned/);
  });

  // 弾く入力: 条件付きの中に、どのテストからも参照されない testid がある -> 検出される
  it('rejects a conditional data-testid that is not referenced by any test or E2E', () => {
    const r = run({
      'src/pages/SamplePage.tsx': `
        export function SamplePage({ showNote }) {
          return (
            <div>
              {showNote && (
                <div data-testid="unreferenced-cond-note">
                  Notice
                </div>
              )}
            </div>
          );
        }
      `,
      'src/pages/SamplePage.test.tsx': `
        it('renders page', () => {
          render(<SamplePage showNote={false} />);
        });
      `,
    });

    expect(r.status).toBe(1);
    expect(r.stderr).toContain('SamplePage.tsx');
    expect(r.stderr).toContain('unreferenced-cond-note');
    expect(r.stderr).toContain('1 unreferenced conditional testid(s) found');
  });

  // 弾いてはいけない入力: 同じ testid がユニットテストから参照されている -> 検出されない
  it('accepts a conditional data-testid when referenced in a unit test', () => {
    const r = run({
      'src/pages/SamplePage.tsx': `
        export function SamplePage({ showNote }) {
          return (
            <div>
              {showNote && (
                <div data-testid="referenced-cond-note">
                  Notice
                </div>
              )}
            </div>
          );
        }
      `,
      'src/pages/SamplePage.test.tsx': `
        it('renders notice when active', () => {
          render(<SamplePage showNote={true} />);
          expect(screen.getByTestId('referenced-cond-note')).toBeInTheDocument();
        });
      `,
    });

    expect(r.status).toBe(0);
    expect(r.stdout).toContain('conditional-testid-coverage: OK');
    expect(r.stdout).toContain('1 conditional testids checked; all referenced');
  });

  // 弾いてはいけない入力: 同じ testid が E2E テストから参照されている -> 検出されない
  it('accepts a conditional data-testid when referenced in an E2E test', () => {
    const r = run({
      'src/pages/SamplePage.tsx': `
        export function SamplePage({ isGameOver }) {
          return (
            <div>
              {isGameOver && (
                <button data-testid="game-over-restart-btn">
                  Restart
                </button>
              )}
            </div>
          );
        }
      `,
      'e2e/gameplay.ts': `
        test('restarts after game over', async ({ page }) => {
          await page.getByTestId('game-over-restart-btn').click();
        });
      `,
    });

    expect(r.status).toBe(0);
    expect(r.stdout).toContain('conditional-testid-coverage: OK');
    expect(r.stdout).toContain('1 conditional testids checked; all referenced');
  });

  // 条件付きでない testid が未参照でも -> 検出されない (対象外)
  it('ignores unreferenced data-testids that are unconditionally rendered', () => {
    const r = run({
      'src/pages/SamplePage.tsx': `
        export function SamplePage() {
          return (
            <div data-testid="unconditional-container">
              <span data-testid="unconditional-label">Title</span>
            </div>
          );
        }
      `,
      'src/pages/SamplePage.test.tsx': `
        it('renders', () => {});
      `,
    });

    expect(r.status).toBe(0);
    expect(r.stdout).toContain('conditional-testid-coverage: OK');
    expect(r.stdout).toContain('0 conditional testids checked');
  });

  // 部分一致トラップ: 部分一致 (dramaha-omaha-hand-name) では完全一致 (dramaha-omaha-hand) が満たされない
  it('requires exact match and rejects partial matches in tests', () => {
    const r = run({
      'src/pages/DramahaPage.tsx': `
        export function DramahaPage({ showHand }) {
          return (
            <div>
              {showHand && (
                <div data-testid="dramaha-omaha-hand">
                  Hand
                </div>
              )}
            </div>
          );
        }
      `,
      'src/pages/DramahaPage.test.tsx': `
        it('tests hand name only', () => {
          expect(screen.getByTestId('dramaha-omaha-hand-name')).toBeInTheDocument();
        });
      `,
    });

    expect(r.status).toBe(1);
    expect(r.stderr).toContain('DramahaPage.tsx');
    expect(r.stderr).toContain('dramaha-omaha-hand');
  });

  // 動的 testid ($ や { を含む式) は対象外
  it('ignores dynamic data-testids containing expressions or template interpolation', () => {
    const r = run({
      'src/pages/SamplePage.tsx': `
        export function SamplePage({ items }) {
          return (
            <div>
              {items.length > 0 && (
                <div>
                  <div data-testid={\`item-\${idx}\`}>dyn1</div>
                  <div data-testid="item-{$id}">dyn2</div>
                </div>
              )}
            </div>
          );
        }
      `,
      'src/pages/SamplePage.test.tsx': `
        it('renders', () => {});
      `,
    });

    expect(r.status).toBe(0);
    expect(r.stdout).toContain('conditional-testid-coverage: OK');
    expect(r.stdout).toContain('0 conditional testids checked');
  });

  // 条件ブロック脱出 ( )} または </ ) 後の要素は条件付きと判定されない
  it('stops backtracking when exiting a conditional block with )} or </', () => {
    const r = run({
      'src/pages/SamplePage.tsx': `
        export function SamplePage({ cond }) {
          return (
            <div>
              {cond && (
                <div>conditional content</div>
              )}
              <div data-testid="subsequent-unconditional-elem">
                always here
              </div>
            </div>
          );
        }
      `,
      'src/pages/SamplePage.test.tsx': `
        it('renders', () => {});
      `,
    });

    expect(r.status).toBe(0);
    expect(r.stdout).toContain('conditional-testid-coverage: OK');
    expect(r.stdout).toContain('0 conditional testids checked');
  });

  // 三項演算子による条件分岐 {cond ? ( ... ) : null} も検出する
  it('detects unreferenced testids inside ternary conditional rendering', () => {
    const r = run({
      'src/pages/SamplePage.tsx': `
        export function SamplePage({ cond }) {
          return (
            <div>
              {cond ? (
                <div data-testid="ternary-unreferenced-elem">
                  content
                </div>
              ) : null}
            </div>
          );
        }
      `,
    });

    expect(r.status).toBe(1);
    expect(r.stderr).toContain('ternary-unreferenced-elem');
  });

  // camelCase の testid が条件付きブロックの中にあり、テストから参照されている場合 -> 検出されない (許容される)
  it('accepts a conditional camelCase data-testid when referenced in a test', () => {
    const r = run({
      'src/pages/SamplePage.tsx': `
        export function SamplePage({ showHint }) {
          return (
            <div>
              {showHint && (
                <div data-testid="mrsMop-hint-live">
                  Hint
                </div>
              )}
            </div>
          );
        }
      `,
      'src/pages/SamplePage.test.tsx': `
        it('renders hint when active', () => {
          render(<SamplePage showHint={true} />);
          expect(screen.getByTestId('mrsMop-hint-live')).toBeInTheDocument();
        });
      `,
    });

    expect(r.status).toBe(0);
    expect(r.stdout).toContain('conditional-testid-coverage: OK');
    expect(r.stdout).toContain('1 conditional testids checked; all referenced');
  });

  // ページ単位の判定: ページ A の条件付き testid をページ B のテストだけが参照している -> 弾かれる
  it('rejects a conditional testid in PageA when only referenced by PageB test', () => {
    const r = run({
      'src/pages/PageA.tsx': `
        export function PageA({ showBanner }) {
          return (
            <div>
              {showBanner && (
                <div data-testid="page-a-banner">Banner</div>
              )}
            </div>
          );
        }
      `,
      'src/pages/PageB.test.tsx': `
        it('references page a banner in PageB test', () => {
          expect(screen.getByTestId('page-a-banner')).toBeInTheDocument();
        });
      `,
    });

    expect(r.status).toBe(1);
    expect(r.stderr).toContain('PageA.tsx');
    expect(r.stderr).toContain('page-a-banner');
  });

  // ページ単位の判定: ページ A の条件付き testid をページ A 自身のテストが参照している -> 弾かれない
  it('accepts a conditional testid in PageA when referenced by PageA own test', () => {
    const r = run({
      'src/pages/PageA.tsx': `
        export function PageA({ showBanner }) {
          return (
            <div>
              {showBanner && (
                <div data-testid="page-a-banner">Banner</div>
              )}
            </div>
          );
        }
      `,
      'src/pages/PageA.test.tsx': `
        it('renders banner when active', () => {
          expect(screen.getByTestId('page-a-banner')).toBeInTheDocument();
        });
      `,
    });

    expect(r.status).toBe(0);
    expect(r.stdout).toContain('conditional-testid-coverage: OK');
    expect(r.stdout).toContain('1 conditional testids checked; all referenced');
  });

  // components は全体集合のまま: src/components/X.tsx の条件付き testid を src/pages/YPage.test.tsx が参照している -> 弾かれない
  it('accepts a conditional testid in src/components when referenced by a page test', () => {
    const r = run({
      'src/components/SharedBadge.tsx': `
        export function SharedBadge({ active }) {
          return (
            <div>
              {active && (
                <span data-testid="shared-status-badge">Active</span>
              )}
            </div>
          );
        }
      `,
      'src/pages/YPage.test.tsx': `
        it('renders shared badge via page test', () => {
          expect(screen.getByTestId('shared-status-badge')).toBeInTheDocument();
        });
      `,
    });

    expect(r.status).toBe(0);
    expect(r.stdout).toContain('conditional-testid-coverage: OK');
    expect(r.stdout).toContain('1 conditional testids checked; all referenced');
  });

  // 遡り幅: 条件が testid の 50 行上にあるフィクスチャで、未参照なら弾かれる (40 のままなら見逃す形)
  it('rejects an unreferenced conditional testid whose condition is 50 lines above', () => {
    const fillers = Array.from({ length: 50 }, (_, k) => `                <p>filler ${k}</p>`).join('\n');
    const r = run({
      'src/pages/DeepPage.tsx': `
        export function DeepPage({ showDeep }) {
          return (
            <div>
              {showDeep && (
                <div>
${fillers}
                  <div data-testid="deep-50-line-testid">deep content</div>
                </div>
              )}
            </div>
          );
        }
      `,
      'src/pages/DeepPage.test.tsx': `
        it('renders deep page', () => {});
      `,
    });

    expect(r.status).toBe(1);
    expect(r.stderr).toContain('DeepPage.tsx');
    expect(r.stderr).toContain('deep-50-line-testid');
  });

  // --root に値が無い: exit 1 になり、TypeError のスタックではなく理由が出ること
  it('exits 1 with an error message and no TypeError stack when --root has no value', () => {
    const r = spawnSync(process.execPath, [GUARD, '--root'], {
      encoding: 'utf8',
      cwd: process.cwd(),
    });

    expect(r.status).toBe(1);
    expect(r.stderr).toContain('--root requires a directory argument');
    expect(r.stderr).not.toContain('TypeError');
  });
});
