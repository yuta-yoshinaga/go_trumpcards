import { describe, expect, it } from 'vitest';
import type { Card, CardDesign } from '../types/card';
import type { FortressTableauCard } from '../types/games/fortress';
import { fortressLegalTargets } from './fortressLegalTargets';

const card = (design: CardDesign, value: number): Card => ({ design, value });
const tc = (design: CardDesign, value: number): FortressTableauCard => ({
  card: card(design, value),
  faceUp: true,
});

describe('fortressLegalTargets', () => {
  it('reports nothing without a selected card', () => {
    const targets = fortressLegalTargets([[tc('SPADE', 5)]], [[]], null);
    expect(targets.tableau.size).toBe(0);
    expect(targets.foundation.size).toBe(0);
  });

  // **Fortress は同スートで隣接ランク（昇順・降順どちらも）。**
  // クローン元の Beleaguered Castle は「スートを見ずに1つ下」で、この関数はその規則の
  // まま出荷されていた (#6187 のレビューで発覚)。合法手でない列を光らせていた。
  it('accepts the same suit one rank higher OR lower, and rejects other suits', () => {
    const targets = fortressLegalTargets(
      [
        [tc('SPADE', 6)], // same suit, one higher -> legal
        [tc('SPADE', 4)], // same suit, one lower  -> legal (Fortress builds both ways)
        [tc('HEART', 6)], // adjacent rank, wrong suit -> NOT legal
        [tc('SPADE', 8)], // same suit, two apart -> NOT legal
      ],
      [[], [], [], []],
      card('SPADE', 5),
    );
    expect([...targets.tableau].sort()).toEqual([0, 1]);
  });

  // **空き列にはどのカードでも置ける。**姉妹の Baker's Dozen は埋められないので、
  // そちらの規則を流用すると実際には打てる手を落とす。
  it('offers an empty column to any card', () => {
    const targets = fortressLegalTargets([[], [tc('SPADE', 9)]], [[], [], [], []], card('HEART', 5));
    expect(targets.tableau.has(0)).toBe(true);
    expect(targets.tableau.has(1)).toBe(false);
  });

  // **ファンデーションは逆に同スート。**タブローの規則を流用すると、置けない
  // 山を光らせる。
  it('requires the same suit going up on a foundation', () => {
    const foundation: Card[][] = [[card('SPADE', 4)], [card('HEART', 4)], [], []];
    const targets = fortressLegalTargets([[tc('CLOVER', 9)]], foundation, card('SPADE', 5));
    expect([...targets.foundation]).toEqual([0]);
  });

  it('starts an empty foundation only with an ace', () => {
    expect(fortressLegalTargets([], [[]], card('SPADE', 1)).foundation.has(0)).toBe(true);
    expect(fortressLegalTargets([], [[]], card('SPADE', 2)).foundation.has(0)).toBe(false);
  });

  it('offers a tableau column and a foundation at once when both accept', () => {
    // Same suit for the tableau (Fortress builds by suit), same suit for the foundation.
    const targets = fortressLegalTargets([[tc('SPADE', 6)]], [[card('SPADE', 4)], [], [], []], card('SPADE', 5));
    expect(targets.tableau.has(0)).toBe(true);
    expect(targets.foundation.has(0)).toBe(true);
  });
});
