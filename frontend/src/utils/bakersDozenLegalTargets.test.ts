import { describe, expect, it } from 'vitest';
import type { BakersDozenTableauCard, Card, CardDesign } from '../types/card';
import { bakersDozenLegalTargets } from './bakersDozenLegalTargets';

const card = (design: CardDesign, value: number): Card => ({ design, value });
const tc = (design: CardDesign, value: number): BakersDozenTableauCard => ({ card: card(design, value), faceUp: true });

describe('bakersDozenLegalTargets', () => {
  it('reports nothing without a selected card', () => {
    const targets = bakersDozenLegalTargets([[tc('SPADE', 5)]], [[]], null);
    expect(targets.tableau.size).toBe(0);
    expect(targets.foundation.size).toBe(0);
  });

  // **タブローはスートを見ない。**ランクが1つ下がるかどうかだけ。
  it('accepts any suit onto a tableau card one rank higher', () => {
    const targets = bakersDozenLegalTargets(
      [[tc('SPADE', 6)], [tc('HEART', 6)], [tc('CLOVER', 9)]],
      [[], [], [], []],
      card('DIAMOND', 5),
    );
    expect([...targets.tableau].sort()).toEqual([0, 1]);
  });

  // **空き列には置けない。**他のソリティアと違い Baker's Dozen は埋められない。
  it('never offers an empty column', () => {
    const targets = bakersDozenLegalTargets([[], [tc('SPADE', 6)]], [[], [], [], []], card('HEART', 5));
    expect(targets.tableau.has(0)).toBe(false);
    expect(targets.tableau.has(1)).toBe(true);
  });

  // **ファンデーションは逆に同スート。**タブローの規則を流用すると、置けない
  // 山を光らせる。
  it('requires the same suit going up on a foundation', () => {
    const foundation: Card[][] = [[card('SPADE', 4)], [card('HEART', 4)], [], []];
    const targets = bakersDozenLegalTargets([[tc('CLOVER', 9)]], foundation, card('SPADE', 5));
    expect([...targets.foundation]).toEqual([0]);
  });

  it('starts an empty foundation only with an ace', () => {
    expect(bakersDozenLegalTargets([], [[]], card('SPADE', 1)).foundation.has(0)).toBe(true);
    expect(bakersDozenLegalTargets([], [[]], card('SPADE', 2)).foundation.has(0)).toBe(false);
  });

  it('offers a tableau column and a foundation at once when both accept', () => {
    const targets = bakersDozenLegalTargets([[tc('CLOVER', 6)]], [[card('SPADE', 4)], [], [], []], card('SPADE', 5));
    expect(targets.tableau.has(0)).toBe(true);
    expect(targets.foundation.has(0)).toBe(true);
  });
});
