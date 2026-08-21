import { describe, expect, it } from 'vitest';
import type { Card, CardDesign } from '../types/card';
import type { SomersetTableauCard } from '../types/games/somerset';
import { somersetLegalTargets } from './somersetLegalTargets';

const card = (design: CardDesign, value: number): Card => ({ design, value });
const tc = (design: CardDesign, value: number): SomersetTableauCard => ({
  card: card(design, value),
  faceUp: true,
});

describe('somersetLegalTargets', () => {
  it('reports nothing without a selected card', () => {
    const targets = somersetLegalTargets([[tc('SPADE', 5)]], [[]], null);
    expect(targets.tableau.size).toBe(0);
    expect(targets.foundation.size).toBe(0);
  });

  // **Somerset は異色で1つ下がるときだけ。**クローン元の Fortress は「同スートで
  // 隣接ランク」で、この関数はその規則のまま出荷されかけた (#6187 のレビューで発覚)。
  it('accepts an alternating colour one rank higher, and rejects same colour / ascending', () => {
    const targets = somersetLegalTargets(
      [
        [tc('HEART', 6)],   // red over black 5, one higher -> legal
        [tc('DIAMOND', 6)], // the other red suit           -> legal
        [tc('CLOVER', 6)],  // same colour as the spade 5   -> NOT legal
        [tc('HEART', 4)],   // alternating but ASCENDING    -> NOT legal
        [tc('HEART', 8)],   // alternating but two apart    -> NOT legal
      ],
      [[], [], [], []],
      card('SPADE', 5),
    );
    expect([...targets.tableau].sort()).toEqual([0, 1]);
  });

  // **空き列にはどのカードでも置ける。**姉妹の Baker's Dozen は埋められないので、
  // そちらの規則を流用すると実際には打てる手を落とす。
  it('offers an empty column to any card', () => {
    const targets = somersetLegalTargets([[], [tc('SPADE', 9)]], [[], [], [], []], card('HEART', 5));
    expect(targets.tableau.has(0)).toBe(true);
    expect(targets.tableau.has(1)).toBe(false);
  });

  // **ファンデーションは逆に同スート。**タブローの規則を流用すると、置けない
  // 山を光らせる。
  it('requires the same suit going up on a foundation', () => {
    const foundation: Card[][] = [[card('SPADE', 4)], [card('HEART', 4)], [], []];
    const targets = somersetLegalTargets([[tc('CLOVER', 9)]], foundation, card('SPADE', 5));
    expect([...targets.foundation]).toEqual([0]);
  });

  it('starts an empty foundation only with an ace', () => {
    expect(somersetLegalTargets([], [[]], card('SPADE', 1)).foundation.has(0)).toBe(true);
    expect(somersetLegalTargets([], [[]], card('SPADE', 2)).foundation.has(0)).toBe(false);
  });

  it('offers a tableau column and a foundation at once when both accept', () => {
    // Alternating colour for the tableau (red 6 takes the black 5), same suit for the foundation.
    const targets = somersetLegalTargets([[tc('HEART', 6)]], [[card('SPADE', 4)], [], [], []], card('SPADE', 5));
    expect(targets.tableau.has(0)).toBe(true);
    expect(targets.foundation.has(0)).toBe(true);
  });
});
