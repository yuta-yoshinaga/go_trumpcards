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
 * 領域を**自己閉じ**で書き、その隣に催促が **1 つだけ** ある形。
 *
 * `siblingRegion` は催促が 2 つあるので 2 つ目が必ず範囲外になり、1 つ目を
 * 取りこぼしても検査が緑にならない ── つまり**この形でしか出ない穴**だった
 * (PR #7042 レビュー指摘)。自己閉じタグは深さを増やさないので、素朴な対応付けは
 * 「次の兄弟の閉じタグ」を領域の終わりだと読んでしまう。
 */
const selfClosingSingle = (slug) => `export function Page() {
  return (
    <div>
      <div data-testid="${slug}-prompt-live" role="status" aria-live="polite" />
      {canBid && (<div data-testid="${slug}-bid-prompt">{t('bidPhase')}</div>)}
    </div>
  );
}
`;

/** flex の行に置く催促のための span 版の領域。 */
const spanRegion = (slug) => `export function Page() {
  return (
    <div className="flex gap-2">
      <span data-testid="${slug}-prompt-live" role="status" aria-live="polite">
        {isDeclarer && (<span data-testid="${slug}-declare-prompt">{t('declarePrompt')}</span>)}
      </span>
    </div>
  );
}
`;

/** bid/discard 以外の名前の催促。**名前で漏れた**のが Tysiac の talon だった。 */
const otherKinds = (slug) => `export function Page() {
  return (
    <div>
      {canDiscard && (<div data-testid="${slug}-talon-prompt">{t('talonPhase')}</div>)}
      {canNameTrump && (<div data-testid="${slug}-trump-prompt">{t('trumpPrompt')}</div>)}
    </div>
  );
}
`;

/**
 * 催促そのものが常設のライブ領域になっている形 (MonteCarlo / FourteenOut)。
 * 中身だけが差し替わるので**これで要件は満たされている** ── 外側にもう 1 枚
 * 包めと言うのは形だけの模倣になる。
 */
const selfLivePrompt = (slug) => `export function Page() {
  return (
    <div>
      <div data-testid="${slug}-prompt" role="status" aria-live="polite">
        {isHumanTurn ? t('turnYours') : t('turnCpu')}
      </div>
    </div>
  );
}
`;

/** 同じ属性でも**条件付きで生える**なら読み上げられない。上の免除の負のコントロール。 */
const conditionalSelfLive = (slug) => `export function Page() {
  return (
    <div>
      {isPlayPhase && (
        <div data-testid="${slug}-prompt" role="status" aria-live="polite">
          {isHumanTurn ? t('turnYours') : t('turnCpu')}
        </div>
      )}
    </div>
  );
}
`;

/**
 * ページ群を書き出してガードを走らせる。
 *
 * 床を越えるだけの準拠ページ (25 枚 = 50 催促) で埋めてから対象を足すので、**落ちた理由が床でなく
 * 検査そのもの**であることが保証される。
 */
function run(extra) {
  const dir = mkdtempSync(join(tmpdir(), 'prompt-live-'));
  dirs.push(dir);
  for (let i = 0; i < 25; i += 1) {
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

  // 自己閉じの領域は子を持てない。素朴な対応付けだと「次の兄弟の閉じ」を
  // 領域の終わりと読み、隣にあるだけの催促を「中にある」と誤判定する。
  it('fails a self-closing region with a single sibling prompt', () => {
    const { code, out } = run({ 'SelfClosePage.tsx': selfClosingSingle('selfclose') });
    expect(code).not.toBe(0);
    expect(out).toMatch(/selfclose-bid-prompt is outside selfclose-prompt-live/);
  });

  it('accepts a span region for a prompt that sits in a flex row', () => {
    const { code, out } = run({ 'SpanPage.tsx': spanRegion('span') });
    expect(code).toBe(0);
    expect(out).toContain('prompt-live-region: OK');
  });

  // **名前が違うだけの催促を見落とさない。** talon/trump は Tysiac と Cinch が
  // 実際に取りこぼしていた名前。
  it('scans talon and trump prompts too, not just bid/discard', () => {
    const { code, out } = run({ 'OtherPage.tsx': otherKinds('other') });
    expect(code).not.toBe(0);
    expect(out).toContain('other-talon-prompt, other-trump-prompt has no *-prompt-live region');
  });

  it('accepts a prompt that is itself an always-mounted live region', () => {
    const { code, out } = run({ 'SelfLivePage.tsx': selfLivePrompt('selflive') });
    expect(code).toBe(0);
    expect(out).toContain('prompt-live-region: OK');
  });

  // 上の免除が「属性さえあれば通る」に化けていないこと。
  it('still fails a self-live prompt that is mounted conditionally', () => {
    const { code, out } = run({ 'CondSelfPage.tsx': conditionalSelfLive('condself') });
    expect(code).not.toBe(0);
    expect(out).toContain('condself-prompt has no *-prompt-live region');
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
