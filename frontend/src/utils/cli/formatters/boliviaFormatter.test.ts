import { describe, expect, it } from 'vitest';
import { makeBoliviaState } from '../../../test/stateFactories';
import { BOLIVIA_MELD_KIND } from '../../../types/games/bolivia';
import { formatBoliviaState } from './boliviaFormatter';

const wildSeven = [
  { design: 'SPADE' as const, value: 2 },
  { design: 'HEART' as const, value: 2 },
  { design: 'CLOVER' as const, value: 2 },
  { design: 'DIAMOND' as const, value: 2 },
  { design: 'JOKER' as const, value: 0 },
  { design: 'JOKER' as const, value: 0 },
  { design: 'JOKER' as const, value: 0 },
];

const withMeld = (meld: Record<string, unknown>) => {
  const base = makeBoliviaState();
  return makeBoliviaState({
    players: base.players.map((p, i) => (i === 0 ? { ...p, melds: [meld as never] } : p)),
  });
};

describe('formatBoliviaState meld labels', () => {
  // **`isBolivia` が真になるのはワイルドのメルドだけ。** kind 1 で分岐して
  // いたころは、完成したボリビアが `(set)` と表示されていた (レビュー指摘)。
  it('names a completed wild meld a bolivia, never a set', () => {
    const out = formatBoliviaState(
      withMeld({
        cards: wildSeven,
        kind: BOLIVIA_MELD_KIND.WILD,
        isNatural: false,
        isCanasta: false,
        isEscalera: false,
        isBolivia: true,
        rank: 0,
      }),
    );
    expect(out).toContain('(bolivia)');
    expect(out).not.toContain('(set)');
    expect(out).not.toContain('(canasta)');
  });

  it('calls an unfinished wild meld a wild, not a bolivia', () => {
    const out = formatBoliviaState(
      withMeld({
        cards: wildSeven.slice(0, 4),
        kind: BOLIVIA_MELD_KIND.WILD,
        isNatural: false,
        isCanasta: false,
        isEscalera: false,
        isBolivia: false,
        rank: 0,
      }),
    );
    expect(out).toContain('(wild)');
    expect(out).not.toContain('(bolivia)');
  });

  it('names a completed sequence an escalera', () => {
    const out = formatBoliviaState(
      withMeld({
        cards: Array.from({ length: 7 }, (_, i) => ({ design: 'HEART' as const, value: 4 + i })),
        kind: BOLIVIA_MELD_KIND.ESCALERA,
        isNatural: true,
        isCanasta: false,
        isEscalera: true,
        isBolivia: false,
        rank: 4,
      }),
    );
    expect(out).toContain('(escalera)');
    expect(out).not.toContain('(bolivia)');
  });

  it('still names sets and canastas', () => {
    const out = formatBoliviaState(
      withMeld({
        cards: Array.from({ length: 7 }, () => ({ design: 'SPADE' as const, value: 5 })),
        kind: BOLIVIA_MELD_KIND.SET,
        isNatural: true,
        isCanasta: true,
        isEscalera: false,
        isBolivia: false,
        rank: 5,
      }),
    );
    expect(out).toContain('(canasta)');
  });

  // **席の行にエスカレラの印が要る。** 上がりを止めているのはこちらで、
  // クローン元は役が 2 種類しか無いのでボリビアの印しか無かった。
  it('tags the escalera on the seat line, distinctly from a bolivia', () => {
    const base = makeBoliviaState();
    const out = formatBoliviaState(
      makeBoliviaState({
        players: base.players.map((p, i) => (i === 0 ? { ...p, hasEscalera: true, hasBolivia: false } : p)),
      }),
    );
    // **席の行だけを見る。** "Bolivia" は見出しにも出るので、出力全体に対する
    // 否定は決して成立しない ── それでは何も測っていない。
    const seat = out.split('\n').find((l) => l.includes('team 0') && l.includes('[')) as string;
    expect(seat).toContain('Escalera');
    expect(seat).not.toContain('Bolivia');
  });

  it('tags a bolivia distinctly from an escalera', () => {
    const base = makeBoliviaState();
    const out = formatBoliviaState(
      makeBoliviaState({
        players: base.players.map((p, i) => (i === 0 ? { ...p, hasEscalera: false, hasBolivia: true } : p)),
      }),
    );
    const seat = out.split('\n').find((l) => l.includes('team 0') && l.includes('[')) as string;
    expect(seat).toContain('Bolivia');
    expect(seat).not.toContain('Escalera');
  });
});
