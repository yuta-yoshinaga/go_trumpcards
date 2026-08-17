import { describe, expect, it } from 'vitest';
import type { AmericanToadTableauCard, Card, CardDesign } from '../types/card';
import { americanToadLegalTargets } from './americanToadLegalTargets';

const card = (design: CardDesign, value: number): Card => ({ design, value }) as Card;
const col = (...cards: Card[]): AmericanToadTableauCard[] => cards.map((c) => ({ card: c, faceUp: true }));

const emptyFoundations = (): Card[][] => Array.from({ length: 8 }, () => []);
const emptyTableau = (): AmericanToadTableauCard[][] => Array.from({ length: 8 }, () => []);

describe('americanToadLegalTargets', () => {
  it('is empty without a card', () => {
    const r = americanToadLegalTargets(emptyTableau(), emptyFoundations(), [], 5, null);
    expect(r.tableau.size).toBe(0);
    expect(r.foundation.size).toBe(0);
  });

  // タブローは**同スートの降順**。ランクだけ合っていてもスートが違えば置けない。
  it('accepts the same suit one rank down on the tableau', () => {
    const tableau = emptyTableau();
    tableau[0] = col(card('SPADE', 8));
    tableau[1] = col(card('HEART', 8));
    const r = americanToadLegalTargets(tableau, emptyFoundations(), [card('SPADE', 2)], 5, card('SPADE', 7));
    expect([...r.tableau]).toEqual([0]);
  });

  // **A の下は K。**普通の 1..13 で書くと、この手が置けない列として表示される。
  it('wraps King under Ace on the tableau', () => {
    const tableau = emptyTableau();
    tableau[3] = col(card('CLOVER', 1));
    const r = americanToadLegalTargets(tableau, emptyFoundations(), [card('SPADE', 2)], 5, card('CLOVER', 13));
    expect([...r.tableau]).toEqual([3]);
  });

  // **リザーブが残っている間、空列は置き先ではない** (自動補充の対象)。
  it('offers an empty column only once the reserve is gone', () => {
    const withReserve = americanToadLegalTargets(
      emptyTableau(),
      emptyFoundations(),
      [card('SPADE', 2)],
      5,
      card('HEART', 4),
    );
    expect(withReserve.tableau.size).toBe(0);

    const noReserve = americanToadLegalTargets(emptyTableau(), emptyFoundations(), [], 5, card('HEART', 4));
    expect(noReserve.tableau.size).toBe(8);
  });

  // 基礎札は**そのインデックスのスート**しか受け取らない。同スートは 2 つある。
  it('accepts the base rank on both foundations of the matching suit', () => {
    const r = americanToadLegalTargets(emptyTableau(), emptyFoundations(), [], 5, card('HEART', 5));
    expect([...r.foundation].sort((a, b) => a - b)).toEqual([2, 6]);
  });

  it('builds a foundation up in suit, wrapping Ace over King', () => {
    const foundation = emptyFoundations();
    foundation[0] = [card('SPADE', 13)];
    const r = americanToadLegalTargets(emptyTableau(), foundation, [], 5, card('SPADE', 1));
    expect(r.foundation.has(0)).toBe(true);
  });

  it('refuses a finished foundation', () => {
    const foundation = emptyFoundations();
    foundation[0] = Array.from({ length: 13 }, (_, i) => card('SPADE', i + 1));
    const r = americanToadLegalTargets(emptyTableau(), foundation, [], 5, card('SPADE', 1));
    expect(r.foundation.has(0)).toBe(false);
  });

  // **基準ランクが決まる前は誰も受け取らない。**
  it('offers no foundation before the base rank is fixed', () => {
    const r = americanToadLegalTargets(emptyTableau(), emptyFoundations(), [], 0, card('HEART', 5));
    expect(r.foundation.size).toBe(0);
  });

  // 完成していない基礎札でも、**続きでない札**は受け取らない。
  it('refuses a card that does not continue the foundation', () => {
    const foundation = emptyFoundations();
    foundation[0] = [card('SPADE', 5)];
    const r = americanToadLegalTargets(emptyTableau(), foundation, [], 5, card('SPADE', 9));
    expect(r.foundation.has(0)).toBe(false);
  });

  // タブローの**ランクは合うがスートが違う**札も受け取らない (両方の枝を通す)。
  it('refuses the right rank in the wrong suit', () => {
    const tableau = emptyTableau();
    tableau[0] = col(card('SPADE', 8));
    // リザーブを残して空列を候補から外し、スート判定だけを見る。
    const r = americanToadLegalTargets(tableau, emptyFoundations(), [card('SPADE', 2)], 5, card('HEART', 7));
    expect(r.tableau.size).toBe(0);
  });
});
