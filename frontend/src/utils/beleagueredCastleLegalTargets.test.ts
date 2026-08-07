import { describe, expect, it } from 'vitest';
import type { Card, CardDesign } from '../types/card';
import type { BeleagueredCastleTableauCard } from '../types/games/beleagueredcastle';
import { beleagueredCastleLegalTargets } from './beleagueredCastleLegalTargets';

const card = (design: CardDesign, value: number): Card => ({ design, value });
const tc = (design: CardDesign, value: number): BeleagueredCastleTableauCard => ({
  card: card(design, value),
  faceUp: true,
});

describe('beleagueredCastleLegalTargets', () => {
  it('reports nothing without a selected card', () => {
    const targets = beleagueredCastleLegalTargets([[tc('SPADE', 5)]], [[]], null);
    expect(targets.tableau.size).toBe(0);
    expect(targets.foundation.size).toBe(0);
  });

  // **タブローはスートを見ない。**ランクが1つ下がるかどうかだけ。
  it('accepts any suit onto a tableau card one rank higher', () => {
    const targets = beleagueredCastleLegalTargets(
      [[tc('SPADE', 6)], [tc('HEART', 6)], [tc('CLOVER', 9)]],
      [[], [], [], []],
      card('DIAMOND', 5),
    );
    expect([...targets.tableau].sort()).toEqual([0, 1]);
  });

  // **空き列にはどのカードでも置ける。**姉妹の Baker's Dozen は埋められないので、
  // そちらの規則を流用すると実際には打てる手を落とす。
  it('offers an empty column to any card', () => {
    const targets = beleagueredCastleLegalTargets([[], [tc('SPADE', 9)]], [[], [], [], []], card('HEART', 5));
    expect(targets.tableau.has(0)).toBe(true);
    expect(targets.tableau.has(1)).toBe(false);
  });

  // **ファンデーションは逆に同スート。**タブローの規則を流用すると、置けない
  // 山を光らせる。
  it('requires the same suit going up on a foundation', () => {
    const foundation: Card[][] = [[card('SPADE', 4)], [card('HEART', 4)], [], []];
    const targets = beleagueredCastleLegalTargets([[tc('CLOVER', 9)]], foundation, card('SPADE', 5));
    expect([...targets.foundation]).toEqual([0]);
  });

  it('starts an empty foundation only with an ace', () => {
    expect(beleagueredCastleLegalTargets([], [[]], card('SPADE', 1)).foundation.has(0)).toBe(true);
    expect(beleagueredCastleLegalTargets([], [[]], card('SPADE', 2)).foundation.has(0)).toBe(false);
  });

  it('offers a tableau column and a foundation at once when both accept', () => {
    const targets = beleagueredCastleLegalTargets(
      [[tc('CLOVER', 6)]],
      [[card('SPADE', 4)], [], [], []],
      card('SPADE', 5),
    );
    expect(targets.tableau.has(0)).toBe(true);
    expect(targets.foundation.has(0)).toBe(true);
  });
});
