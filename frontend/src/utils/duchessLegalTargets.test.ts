import { describe, expect, it } from 'vitest';
import type { Card, CardDesign, DuchessTableauCard } from '../types/card';
import { duchessLegalTargets } from './duchessLegalTargets';

const card = (design: CardDesign, value: number): Card => ({ design, value });
const tableau = (...columns: DuchessTableauCard[][]): DuchessTableauCard[][] =>
  Array.from({ length: 4 }, (_, index) => columns[index] ?? []);
const column = (...cards: Card[]): DuchessTableauCard[] => cards.map((value) => ({ card: value, faceUp: true }));
const foundations = (): Card[][] => Array.from({ length: 4 }, () => []);

describe('duchessLegalTargets', () => {
  it('accepts only a different-colour card one rank down on the tableau', () => {
    const result = duchessLegalTargets(
      tableau(
        column(card('HEART', 10)),
        column(card('DIAMOND', 10)),
        column(card('CLOVER', 10)),
        column(card('HEART', 11)),
      ),
      foundations(),
      [[], [], [], []],
      5,
      false,
      card('SPADE', 9),
    );
    expect([...result.tableau]).toEqual([0, 1]);
    expect(result.tableau.has(2)).toBe(false);
  });

  it('wraps Ace below King and King above Ace', () => {
    const belowKing = duchessLegalTargets(
      tableau(column(card('HEART', 13))),
      foundations(),
      [[], [], [], []],
      5,
      false,
      card('CLOVER', 12),
    );
    const belowAce = duchessLegalTargets(
      tableau(column(card('CLOVER', 1))),
      foundations(),
      [[], [], [], []],
      5,
      false,
      card('HEART', 13),
    );
    expect(belowKing.tableau.has(0)).toBe(true);
    expect(belowAce.tableau.has(0)).toBe(true);
  });

  it('allows an empty column only after the reserve is exhausted for a tableau source', () => {
    const reserveRemains = duchessLegalTargets(
      tableau([], [], [], []),
      foundations(),
      [[card('SPADE', 2)], [], [], []],
      5,
      false,
      card('HEART', 4),
      'tableau',
    );
    const reserveGone = duchessLegalTargets(
      tableau([], [], [], []),
      foundations(),
      [[], [], [], []],
      5,
      false,
      card('HEART', 4),
      'tableau',
    );
    expect(reserveRemains.tableau.size).toBe(0);
    expect(reserveGone.tableau.size).toBe(4);
  });

  it('allows a reserve source to fill an empty column while reserve remains', () => {
    const result = duchessLegalTargets(
      tableau([], [], [], []),
      foundations(),
      [[card('SPADE', 2)], [], [], []],
      5,
      false,
      card('HEART', 4),
      'reserve',
    );
    expect(result.tableau.size).toBe(4);
  });

  // ドメインは MoveWasteToTableau (Duchess.go:299) でも
  // duchess.errEmptyColumnReserveOnly を返す。空き列はリザーブ専用であって
  // 「タブロー以外なら可」ではない。
  it('shuts a waste source out of an empty column while reserve remains', () => {
    const reserveRemains = duchessLegalTargets(
      tableau([], [], [], []),
      foundations(),
      [[card('SPADE', 2)], [], [], []],
      5,
      false,
      card('HEART', 4),
      'waste',
    );
    const reserveGone = duchessLegalTargets(
      tableau([], [], [], []),
      foundations(),
      [[], [], [], []],
      5,
      false,
      card('HEART', 4),
      'waste',
    );
    expect(reserveRemains.tableau.size).toBe(0);
    expect(reserveGone.tableau.size).toBe(4);
  });

  // Duchess.requireBaseChosen は移動 7 箇所すべての入口にある。開始ランク未決定なら
  // タブローの置き先もひとつも無い。組札だけを黙らせると、配りによっては
  // 置ける先があるように見えて、押すとサーバに拒まれる。
  it('offers no tableau target either before the base rank is chosen', () => {
    const board = tableau(column(card('SPADE', 9)), column(card('HEART', 10)));
    const playable = duchessLegalTargets(board, foundations(), [[], [], [], []], 5, false, card('HEART', 9), 'waste');
    const awaiting = duchessLegalTargets(board, foundations(), [[], [], [], []], 0, true, card('HEART', 9), 'waste');
    // 同じ盤面・同じ札で、開始ランクが決まっていれば置き先はある。
    expect(playable.tableau.size).toBeGreaterThan(0);
    expect(awaiting.tableau.size).toBe(0);
    expect(awaiting.foundation.size).toBe(0);
  });

  it('accepts the base rank on an empty matching foundation', () => {
    const result = duchessLegalTargets(tableau(), foundations(), [[], [], [], []], 5, false, card('HEART', 5));
    expect([...result.foundation]).toEqual([2]);
  });

  it('builds foundations in suit and wraps Ace after King', () => {
    const foundation = foundations();
    foundation[0] = [card('SPADE', 13)];
    const result = duchessLegalTargets(tableau(), foundation, [[], [], [], []], 5, false, card('SPADE', 1));
    expect(result.foundation.has(0)).toBe(true);
  });

  it('offers no foundation before the base rank is available', () => {
    const beforeBase = duchessLegalTargets(tableau(), foundations(), [[], [], [], []], 0, false, card('HEART', 5));
    const awaiting = duchessLegalTargets(tableau(), foundations(), [[], [], [], []], 5, true, card('HEART', 5));
    expect(beforeBase.foundation.size).toBe(0);
    expect(awaiting.foundation.size).toBe(0);
  });

  it('rejects the wrong foundation suit and a finished foundation', () => {
    const wrongSuit = duchessLegalTargets(tableau(), foundations(), [[], [], [], []], 5, false, card('HEART', 5));
    const foundation = foundations();
    foundation[0] = Array.from({ length: 13 }, () => card('SPADE', 1));
    const finished = duchessLegalTargets(tableau(), foundation, [[], [], [], []], 5, false, card('SPADE', 1));
    expect(wrongSuit.foundation.has(0)).toBe(false);
    expect([...wrongSuit.foundation]).toEqual([2]);
    expect(finished.foundation.has(0)).toBe(false);
  });
});
