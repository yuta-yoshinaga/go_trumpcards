import { spawnSync } from 'node:child_process';
import { existsSync, mkdtempSync, rmSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { afterAll, describe, expect, it } from 'vitest';

// vitest はテストモジュールを file: 以外の URL で配るので cwd (= frontend/) から
// 解決し、存在を先に確かめる。パスがずれると全ケースが「無いファイルを spawn した」
// だけになり、落ちるべき入力でも落ちなくなる。
const GUARD = join(process.cwd(), 'scripts', 'check-prompt-live-region.mjs');
if (!existsSync(GUARD)) throw new Error(`check-prompt-live-region.mjs not found at ${GUARD} (cwd: ${process.cwd()})`);

const dirs = [];
afterAll(() => {
  for (const d of dirs) rmSync(d, { recursive: true, force: true });
});

/** 催promptが常設のライブ領域の中にある、正しい形。 */
const compliant = (slug) => `export function Page() {
  return (
    <div data-testid="${slug}-prompt-live" role="status" aria-live="polite">
      {canBid && (
        <div data-testid="${slug}-bid-prompt">{t('bidPhase')}</div>
      )}
      {canDiscard && (
        <div data-testid="${slug}-discard-prompt">{t('discardPhase')}</div>
      )}
    </div>
  );
}
`;

/** 領域が無い。 */
const bare = (slug) => `export function Page() {
  return (
    <div>
      {canBid && (<div data-testid="${slug}-bid-prompt">{t('bidPhase')}</div>)}
      {canDiscard && (<div data-testid="${slug}-discard-prompt">{t('discardPhase')}</div>)}
    </div>
  );
}
`;

/** 領域が条件付きで生える ── 出現と同時のテキストは読み上げられない。 */
const conditionalRegion = (slug) => `export function Page() {
  return (
    <div>
      {canBid && (
        <div data-testid="${slug}-prompt-live" role="status" aria-live="polite">
          <div data-testid="${slug}-bid-prompt">{t('bidPhase')}</div>
          <div data-testid="${slug}-discard-prompt">{t('discardPhase')}</div>
        </div>
      )}
    </div>
  );
}
`;

/** 領域は常設だが催促がその**隣**にある ── 属性の存在だけを見ると通ってしまう形。 */
const siblingRegion = (slug) => `export function Page() {
  return (
    <div>
      <div data-testid="${slug}-prompt-live" role="status" aria-live="polite" />
      {canBid && (<div data-testid="${slug}-bid-prompt">{t('bidPhase')}</div>)}
      {canDiscard && (<div data-testid="${slug}-discard-prompt">{t('discardPhase')}</div>)}
    </div>
  );
}
`;

/** role はあるが aria-live が無い。 */
const noAriaLive = (slug) => `export function Page() {
  return (
    <div data-testid="${slug}-prompt-live" role="status">
      {canBid && (<div data-testid="${slug}-bid-prompt">{t('bidPhase')}</div>)}
      {canDiscard && (<div data-testid="${slug}-discard-prompt">{t('discardPhase')}</div>)}
    </div>
  );
}
`;

/**
 * ページ群を書き出してガードを走らせる。
 *
 * 床を越えるだけの準拠ページで埋めてから対象を足すので、**落ちた理由が床でなく
 * 検査そのもの**であることが保証される。
 */
function run(extra) {
  const dir = mkdtempSync(join(tmpdir(), 'prompt-live-'));
  dirs.push(dir);
  for (let i = 0; i < 9; i += 1) {
    writeFileSync(join(dir, `Filler${i}Page.tsx`), compliant(`filler${i}`));
  }
  for (const [name, body] of Object.entries(extra ?? {})) {
    writeFileSync(join(dir, name), body);
  }
  const r = spawnSync('bun', [GUARD, dir], { encoding: 'utf8' });
  return { code: r.status, out: `${r.stdout}${r.stderr}` };
}

describe('check-prompt-live-region', () => {
  it('passes when every prompt sits inside an always-mounted region', () => {
    const { code, out } = run({ 'GoodPage.tsx': compliant('good') });
    expect(code).toBe(0);
    expect(out).toContain('prompt-live-region: OK');
  });

  it('fails a prompt with no region at all', () => {
    const { code, out } = run({ 'BarePage.tsx': bare('bare') });
    expect(code).not.toBe(0);
    expect(out).toContain('has no *-prompt-live region');
  });

  // **これがこのガードの本題。** 属性の存在だけを見る検査はこの形を通してしまう。
  it('fails a region that is mounted conditionally', () => {
    const { code, out } = run({ 'CondPage.tsx': conditionalRegion('cond') });
    expect(code).not.toBe(0);
    expect(out).toContain('mounted conditionally');
  });

  it('fails a region that merely sits beside the prompt', () => {
    const { code, out } = run({ 'SiblingPage.tsx': siblingRegion('sibling') });
    expect(code).not.toBe(0);
    expect(out).toMatch(/is outside .*, not inside it/);
  });

  it('fails a region missing aria-live', () => {
    const { code, out } = run({ 'NoLivePage.tsx': noAriaLive('nolive') });
    expect(code).not.toBe(0);
    expect(out).toContain('needs both role="status" and aria-live');
  });

  // **0 件で緑にならないこと。** 走査が壊れると「違反なし」に見えてしまう。
  it('fails when the walk finds nothing', () => {
    const dir = mkdtempSync(join(tmpdir(), 'prompt-live-empty-'));
    dirs.push(dir);
    const r = spawnSync('bun', [GUARD, dir], { encoding: 'utf8' });
    expect(r.status).not.toBe(0);
    expect(`${r.stdout}${r.stderr}`).toMatch(/floor|prompt-live-region/i);
  });
});
