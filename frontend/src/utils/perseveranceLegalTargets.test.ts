import { describe, expect, it } from 'vitest';
import type { Card, CardDesign, PerseveranceTableauCard } from '../types/card';
import { perseveranceLegalTargets, perseveranceStartsRun } from './perseveranceLegalTargets';

const card = (design: CardDesign, value: number): Card => ({ design, value });
const tc = (design: CardDesign, value: number): PerseveranceTableauCard => ({
  card: card(design, value),
  faceUp: true,
});

describe('perseveranceLegalTargets', () => {
  it('reports nothing without a selected card', () => {
    const targets = perseveranceLegalTargets([[tc('SPADE', 5)]], [[]], null);
    expect(targets.tableau.size).toBe(0);
    expect(targets.foundation.size).toBe(0);
  });

  // **タブローも同スート。**クローン元の Baker's Dozen はランクだけを見るので、
  // その版のままだと ♠6 も ♥6 も行き先として光ってしまう。
  it('accepts only the same suit onto a tableau card one rank higher', () => {
    const targets = perseveranceLegalTargets(
      [[tc('SPADE', 6)], [tc('HEART', 6)], [tc('CLOVER', 9)]],
      [[], [], [], []],
      card('HEART', 5),
    );
    expect([...targets.tableau]).toEqual([1]);
  });

  // 負のコントロール: ランクが合ってもスートが違えば 1 つも返らない。
  it('refuses a descending move that crosses suit', () => {
    const targets = perseveranceLegalTargets(
      [[tc('SPADE', 6)], [tc('HEART', 6)]],
      [[], [], [], []],
      card('DIAMOND', 5),
    );
    expect(targets.tableau.size).toBe(0);
  });

  // **空き列には置けない。**他のソリティアと違い Perseverance は埋められない。
  it('never offers an empty column', () => {
    const targets = perseveranceLegalTargets([[], [tc('SPADE', 6)]], [[], [], [], []], card('SPADE', 5));
    expect(targets.tableau.has(0)).toBe(false);
    expect(targets.tableau.has(1)).toBe(true);
  });

  // **ファンデーションは逆に同スート。**タブローの規則を流用すると、置けない
  // 山を光らせる。
  it('requires the same suit going up on a foundation', () => {
    const foundation: Card[][] = [[card('SPADE', 4)], [card('HEART', 4)], [], []];
    const targets = perseveranceLegalTargets([[tc('CLOVER', 9)]], foundation, card('SPADE', 5));
    expect([...targets.foundation]).toEqual([0]);
  });

  it('starts an empty foundation only with an ace', () => {
    expect(perseveranceLegalTargets([], [[]], card('SPADE', 1)).foundation.has(0)).toBe(true);
    expect(perseveranceLegalTargets([], [[]], card('SPADE', 2)).foundation.has(0)).toBe(false);
  });

  it('offers a tableau column and a foundation at once when both accept', () => {
    // 卓の行き先も同スートでなければならないので ♠6 を置く (クローン元は ♣6 でも通した)。
    const targets = perseveranceLegalTargets([[tc('SPADE', 6)]], [[card('SPADE', 4)], [], [], []], card('SPADE', 5));
    expect(targets.tableau.has(0)).toBe(true);
    expect(targets.foundation.has(0)).toBe(true);
  });
});

describe('perseveranceStartsRun', () => {
  // ♥K ♠9 ♠8 ♠7 — 上3枚が並び、♥K で切れる。
  const col: PerseveranceTableauCard[] = [tc('HEART', 13), tc('SPADE', 9), tc('SPADE', 8), tc('SPADE', 7)];

  it('accepts the top card and every index the run reaches', () => {
    expect(perseveranceStartsRun(col, 3)).toBe(true);
    expect(perseveranceStartsRun(col, 2)).toBe(true);
    expect(perseveranceStartsRun(col, 1)).toBe(true);
  });

  // 負のコントロール: 並びが切れた下の札は掴めない。
  it('refuses an index below the break', () => {
    expect(perseveranceStartsRun(col, 0)).toBe(false);
  });

  it('refuses a group whose suits differ', () => {
    expect(perseveranceStartsRun([tc('DIAMOND', 8), tc('SPADE', 7)], 0)).toBe(false);
  });

  it('refuses a group that skips a rank', () => {
    expect(perseveranceStartsRun([tc('SPADE', 8), tc('SPADE', 6)], 0)).toBe(false);
  });

  it('refuses an out-of-range index', () => {
    expect(perseveranceStartsRun(col, -1)).toBe(false);
    expect(perseveranceStartsRun(col, 99)).toBe(false);
  });
});
