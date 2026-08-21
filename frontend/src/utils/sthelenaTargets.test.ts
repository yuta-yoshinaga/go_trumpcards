import { describe, expect, it } from 'vitest';
import type { Card, CardDesign, StHelenaTableauCard } from '../types/card';
import { stHelenaCanPlaceOnFoundation, stHelenaCanPlaceOnTableau, stHelenaColumnCanReach } from './sthelenaTargets';

const card = (design: CardDesign, value: number): Card => ({ design, value });
const tc = (design: CardDesign, value: number): StHelenaTableauCard => ({ card: card(design, value), faceUp: true });

/** Foundations pre-seeded like a real deal: 0..3 asc (A), 4..7 desc (K). */
const seededFoundation = (): Card[][] => [
  [card('SPADE', 1)],
  [card('CLOVER', 1)],
  [card('HEART', 1)],
  [card('DIAMOND', 1)],
  [card('SPADE', 13)],
  [card('CLOVER', 13)],
  [card('HEART', 13)],
  [card('DIAMOND', 13)],
];

describe('stHelenaCanPlaceOnFoundation', () => {
  it('accepts the next rank up on an ascending pile of the matching suit', () => {
    expect(stHelenaCanPlaceOnFoundation(card('SPADE', 2), seededFoundation(), 0)).toBe(true);
  });

  it('rejects a non-sequential rank on an ascending pile', () => {
    expect(stHelenaCanPlaceOnFoundation(card('SPADE', 4), seededFoundation(), 0)).toBe(false);
  });

  it('accepts the next rank down on a descending pile of the matching suit', () => {
    expect(stHelenaCanPlaceOnFoundation(card('SPADE', 12), seededFoundation(), 4)).toBe(true);
  });

  it('rejects a suit mismatch even when the rank fits', () => {
    expect(stHelenaCanPlaceOnFoundation(card('HEART', 2), seededFoundation(), 0)).toBe(false);
  });

  it('accepts an Ace onto an empty ascending pile and a King onto an empty descending pile', () => {
    const empty: Card[][] = [[], [], [], [], [], [], [], []];
    expect(stHelenaCanPlaceOnFoundation(card('SPADE', 1), empty, 0)).toBe(true);
    expect(stHelenaCanPlaceOnFoundation(card('SPADE', 2), empty, 0)).toBe(false);
    expect(stHelenaCanPlaceOnFoundation(card('SPADE', 13), empty, 4)).toBe(true);
    expect(stHelenaCanPlaceOnFoundation(card('SPADE', 12), empty, 4)).toBe(false);
  });

  it('rejects out-of-range indices', () => {
    expect(stHelenaCanPlaceOnFoundation(card('SPADE', 2), seededFoundation(), -1)).toBe(false);
    expect(stHelenaCanPlaceOnFoundation(card('SPADE', 2), seededFoundation(), 8)).toBe(false);
  });
});

describe('stHelenaCanPlaceOnTableau', () => {
  const tableau: StHelenaTableauCard[][] = [[tc('SPADE', 5)], [tc('SPADE', 3)], [tc('HEART', 6)], []];

  it('accepts a same-suit card one rank higher', () => {
    expect(stHelenaCanPlaceOnTableau(card('SPADE', 6), tableau, 0)).toBe(true);
  });

  it('accepts a same-suit card one rank lower', () => {
    expect(stHelenaCanPlaceOnTableau(card('SPADE', 4), tableau, 0)).toBe(true);
  });

  it('rejects a same-suit card two ranks away', () => {
    expect(stHelenaCanPlaceOnTableau(card('SPADE', 7), tableau, 0)).toBe(false);
  });

  // **スートは見ない。**元のサブテストは "rejects a suit mismatch" を主張して
  // いたが、それはクローン元クレセントの規則。ここでは通る。
  it('accepts a card of a different suit', () => {
    expect(stHelenaCanPlaceOnTableau(card('HEART', 4), tableau, 0)).toBe(true);
    expect(stHelenaCanPlaceOnTableau(card('DIAMOND', 6), tableau, 0)).toBe(true);
  });

  // **折り返しは無い。**元のサブテストは両方向の成功を主張していた。
  it('rejects the A-K wrap in both directions', () => {
    const wrap: StHelenaTableauCard[][] = [[tc('SPADE', 13)], [tc('SPADE', 1)]];
    expect(stHelenaCanPlaceOnTableau(card('SPADE', 1), wrap, 0)).toBe(false);
    expect(stHelenaCanPlaceOnTableau(card('SPADE', 13), wrap, 1)).toBe(false);
    // 負のコントロール: 折り返しを消したせいで ±1 まで壊していないこと。
    expect(stHelenaCanPlaceOnTableau(card('HEART', 12), wrap, 0)).toBe(true);
  });

  it('rejects an empty column and out-of-range indices', () => {
    expect(stHelenaCanPlaceOnTableau(card('SPADE', 6), tableau, 3)).toBe(false);
    expect(stHelenaCanPlaceOnTableau(card('SPADE', 6), tableau, -1)).toBe(false);
    expect(stHelenaCanPlaceOnTableau(card('SPADE', 6), tableau, 99)).toBe(false);
  });
});

// **どの列がどの段に届くかは、置けるかどうかとは別の判定。**片方だけ実装すると、
// ランクは合っているのに送れない組札を押せる盤になる。
describe('stHelenaColumnCanReach', () => {
  const ACE_ROW = 0;
  const KING_ROW = 4;

  it('lets a top column reach only the king row while restricted', () => {
    for (const col of [0, 1, 2, 3]) {
      expect(stHelenaColumnCanReach(col, KING_ROW, true)).toBe(true);
      expect(stHelenaColumnCanReach(col, ACE_ROW, true)).toBe(false);
    }
  });

  it('lets a bottom column reach only the ace row while restricted', () => {
    for (const col of [6, 7, 8, 9]) {
      expect(stHelenaColumnCanReach(col, ACE_ROW, true)).toBe(true);
      expect(stHelenaColumnCanReach(col, KING_ROW, true)).toBe(false);
    }
  });

  it('lets a side column reach both rows while restricted', () => {
    for (const col of [4, 5, 10, 11]) {
      expect(stHelenaColumnCanReach(col, ACE_ROW, true)).toBe(true);
      expect(stHelenaColumnCanReach(col, KING_ROW, true)).toBe(true);
    }
  });

  // 負のコントロール: 制限が解ければ全部通る。解けても false のままだと、
  // 上のテストは通るのに後半どこにも送れなくなる。
  it('lets every column reach every row once the restriction is lifted', () => {
    for (let col = 0; col < 12; col++) {
      expect(stHelenaColumnCanReach(col, ACE_ROW, false)).toBe(true);
      expect(stHelenaColumnCanReach(col, KING_ROW, false)).toBe(true);
    }
  });
});
