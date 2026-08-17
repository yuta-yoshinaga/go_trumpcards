import { describe, expect, it } from 'vitest';
import type { Card, CardDesign } from '../types/card';
import type { KingAlbertTableauCard } from '../types/games/kingalbert';
import { kingAlbertLegalTargets } from './kingAlbertLegalTargets';

const card = (design: CardDesign, value: number): Card => ({ design, value });
const tc = (design: CardDesign, value: number): KingAlbertTableauCard => ({ card: card(design, value), faceUp: true });

describe('kingAlbertLegalTargets', () => {
  it('accepts only the alternate-colour column one rank higher', () => {
    const tableau: KingAlbertTableauCard[][] = [
      [tc('HEART', 8)], // 赤8 ← 黒7 は置ける
      [tc('SPADE', 8)], // 黒8 ← 黒7 は色が同じなので置けない
      [tc('HEART', 9)], // ランクが合わない
    ];
    const targets = kingAlbertLegalTargets(tableau, [[], [], [], []], card('CLOVER', 7));
    expect([...targets.tableau]).toEqual([0]);
  });

  it('accepts any card on an empty column', () => {
    const targets = kingAlbertLegalTargets([[], [tc('SPADE', 2)]], [[], [], [], []], card('HEART', 13));
    expect(targets.tableau.has(0)).toBe(true);
    expect(targets.tableau.has(1)).toBe(false);
  });

  // 自分の列を明示的に除く必要が無いことの確認 ── 一番上の札と自分自身を比べるので、
  // ランクの条件が必ず偽になる。除外処理を足すと、テストの無い分岐が増えるだけになる。
  it('does not offer the column the card is sitting on', () => {
    const tableau: KingAlbertTableauCard[][] = [[tc('HEART', 8), tc('CLOVER', 7)], []];
    const targets = kingAlbertLegalTargets(tableau, [[], [], [], []], card('CLOVER', 7));
    expect(targets.tableau.has(0)).toBe(false);
    expect(targets.tableau.has(1)).toBe(true);
  });

  it('takes an ace onto an empty foundation and nothing else', () => {
    const empty = kingAlbertLegalTargets([], [[], [], [], []], card('SPADE', 1));
    expect(empty.foundation.size).toBe(4);
    const two = kingAlbertLegalTargets([], [[], [], [], []], card('SPADE', 2));
    expect(two.foundation.size).toBe(0);
  });

  it('builds a foundation up in the same suit only', () => {
    const foundation: Card[][] = [[card('SPADE', 1)], [card('HEART', 1)], [], []];
    const targets = kingAlbertLegalTargets([], foundation, card('SPADE', 2));
    expect([...targets.foundation]).toEqual([0]);
  });

  it('returns nothing when no card is selected', () => {
    const targets = kingAlbertLegalTargets([[]], [[]], null);
    expect(targets.tableau.size).toBe(0);
    expect(targets.foundation.size).toBe(0);
  });
});
