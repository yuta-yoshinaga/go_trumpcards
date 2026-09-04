import { describe, expect, it } from 'vitest';
import type { Card, CardDesign } from '../types/card';
import type { FlowerGardenTableauCard } from '../types/games/flowergarden';
import { flowerGardenLegalTargets } from './flowerGardenLegalTargets';

const card = (design: CardDesign, value: number): Card => ({ design, value });
const tc = (design: CardDesign, value: number): FlowerGardenTableauCard => ({
  card: card(design, value),
  faceUp: true,
});

describe('flowerGardenLegalTargets', () => {
  it('accepts any card on an empty tableau column', () => {
    const tableau: FlowerGardenTableauCard[][] = [[], [tc('SPADE', 2)]];
    const targets = flowerGardenLegalTargets(tableau, [[], [], [], []], card('HEART', 13));
    expect(targets.tableau.has(0)).toBe(true);
    expect(targets.tableau.has(1)).toBe(false);
  });

  it('accepts a column one rank higher regardless of suit or colour (same colour and different colour)', () => {
    // Flower Garden はスート・色を見ない（赤黒交互ではない）。値差 -1 のみ判定する。
    const tableau: FlowerGardenTableauCard[][] = [
      [tc('HEART', 8)], // 赤8 ← 黒7 (異色・異スート): 合法
      [tc('SPADE', 8)], // 黒8 ← 黒7 (同色・異スート): 合法 (King Albert では不可だが Flower Garden では合法)
      [tc('CLOVER', 8)], // 黒8 ← 黒7 (同色・同スート): 合法
      [tc('DIAMOND', 8)], // 赤8 ← 黒7 (異色・異スート): 合法
    ];
    const targets = flowerGardenLegalTargets(tableau, [[], [], [], []], card('CLOVER', 7));
    expect([...targets.tableau]).toEqual([0, 1, 2, 3]);
  });

  it('rejects a tableau column when rank difference is not -1', () => {
    const tableau: FlowerGardenTableauCard[][] = [
      [tc('HEART', 9)], // 値差 -2
      [tc('SPADE', 7)], // 同値
      [tc('CLOVER', 6)], // 値差 +1
      [tc('DIAMOND', 13)], // 任意の値
    ];
    const targets = flowerGardenLegalTargets(tableau, [[], [], [], []], card('CLOVER', 7));
    expect(targets.tableau.size).toBe(0);
  });

  it('does not offer the column the card is sitting on', () => {
    const tableau: FlowerGardenTableauCard[][] = [[tc('HEART', 8), tc('CLOVER', 7)], []];
    const targets = flowerGardenLegalTargets(tableau, [[], [], [], []], card('CLOVER', 7));
    expect(targets.tableau.has(0)).toBe(false);
    expect(targets.tableau.has(1)).toBe(true);
  });

  it('returns empty sets when card is null or undefined', () => {
    const nullTargets = flowerGardenLegalTargets([[]], [[]], null);
    expect(nullTargets.tableau.size).toBe(0);
    expect(nullTargets.foundation.size).toBe(0);

    const undefinedTargets = flowerGardenLegalTargets([[]], [[]], undefined);
    expect(undefinedTargets.tableau.size).toBe(0);
    expect(undefinedTargets.foundation.size).toBe(0);
  });

  it('returns only the first empty foundation when an ace is passed with multiple empty foundations', () => {
    // WebController は index を受け取らず findFoundation で最初の受け入れ可能山に着地するため、
    // 空のファンデーションが複数あっても着地する最初の 1 つだけを返す。
    const emptyFoundations: Card[][] = [[], [], [], []];
    const targets = flowerGardenLegalTargets([], emptyFoundations, card('SPADE', 1));
    expect(targets.foundation.size).toBe(1);
    expect([...targets.foundation]).toEqual([0]);

    const partialFoundations: Card[][] = [[card('HEART', 1)], [], [], []];
    const aceTargets = flowerGardenLegalTargets([], partialFoundations, card('SPADE', 1));
    expect(aceTargets.foundation.size).toBe(1);
    expect([...aceTargets.foundation]).toEqual([1]);

    const nonAceTargets = flowerGardenLegalTargets([], emptyFoundations, card('SPADE', 2));
    expect(nonAceTargets.foundation.size).toBe(0);
  });

  it('builds up a foundation in the same suit only and rejects different suits', () => {
    const foundation: Card[][] = [[card('SPADE', 1)], [card('HEART', 1)], [card('CLOVER', 5)], []];
    // ♠2 は ♠1 (idx 0) に置けるが ♥1 (idx 1) や ♣5 (idx 2) には置けない
    const spade2 = flowerGardenLegalTargets([], foundation, card('SPADE', 2));
    expect([...spade2.foundation]).toEqual([0]);

    // ♣6 は ♣5 (idx 2) に置ける
    const clover6 = flowerGardenLegalTargets([], foundation, card('CLOVER', 6));
    expect([...clover6.foundation]).toEqual([2]);

    // ♦2 は空ファンデーション (idx 3) には置けない (A ではない)
    const diamond2 = flowerGardenLegalTargets([], foundation, card('DIAMOND', 2));
    expect(diamond2.foundation.size).toBe(0);
  });
});
