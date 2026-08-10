import { describe, expect, it } from 'vitest';
import type { TriPeaksCard } from '../types/games/tripeaks';
import { triPeaksPlayableCells } from './triPeaksPlayable';

const cell = (value: number, over: Partial<TriPeaksCard> = {}): TriPeaksCard => ({
  card: { design: 'SPADE', value },
  removed: false,
  exposed: true,
  ...over,
});

describe('triPeaksPlayableCells', () => {
  it('reports nothing while the waste is empty', () => {
    expect(triPeaksPlayableCells([[cell(5)]], undefined).size).toBe(0);
  });

  it('counts only the exposed cards one rank away from the waste top', () => {
    const layout = [[cell(6), cell(4), cell(9)]];
    expect([...triPeaksPlayableCells(layout, 5)]).toEqual(['0-0', '0-1']);
  });

  // 伏せ札・除去済みは触れない。除外しないと CUI の playableCount とずれる。
  it('skips cards that are removed or not yet exposed', () => {
    const layout = [[cell(6, { exposed: false }), cell(4, { removed: true }), cell(6)]];
    expect([...triPeaksPlayableCells(layout, 5)]).toEqual(['0-2']);
  });

  // K↔A は隣接する。ここを落とすと「出せるのに0枚」と表示される。
  it('wraps between king and ace', () => {
    expect(triPeaksPlayableCells([[cell(13)]], 1).size).toBe(1);
    expect(triPeaksPlayableCells([[cell(1)]], 13).size).toBe(1);
  });

  it('reports the row and column of each playable cell', () => {
    const layout = [[cell(9)], [cell(2), cell(4)]];
    expect([...triPeaksPlayableCells(layout, 3)]).toEqual(['1-0', '1-1']);
  });
});
