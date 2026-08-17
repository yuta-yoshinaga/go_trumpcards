import { describe, expect, it } from 'vitest';
import type { AmericanToadTableauCard, Card, CardDesign } from '../types/card';
import { americanToadLegalTargets, americanToadSourceCard } from './americanToadLegalTargets';

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

  // **空列はタブロー同士の組み替えでは埋められない** — リザーブと捨て札の出口
  // だから (`MoveTableauToTableau`, #4417)。出どころを渡さないと表現できない。
  it('never offers an empty column to a card coming from another column', () => {
    const fromTableau = americanToadLegalTargets(
      emptyTableau(),
      emptyFoundations(),
      [],
      5,
      card('HEART', 4),
      'tableau',
    );
    expect(fromTableau.tableau.size).toBe(0);

    // 同じ盤面・同じ札でも、リザーブと捨て札からなら置ける。
    for (const zone of ['reserve', 'waste']) {
      const r = americanToadLegalTargets(emptyTableau(), emptyFoundations(), [], 5, card('HEART', 4), zone);
      expect(r.tableau.size).toBe(8);
    }
  });

  // 出どころの制限は空列だけの話。埋まっている列は変わらず判定する。
  it('still offers a matching non-empty column to a tableau card', () => {
    const tableau = emptyTableau();
    tableau[2] = col(card('HEART', 5));
    const r = americanToadLegalTargets(tableau, emptyFoundations(), [], 5, card('HEART', 4), 'tableau');
    expect([...r.tableau]).toEqual([2]);
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

// 光らせる対象を決める側。ゾーンが何も指していないときに undefined を返すのが
// 肝で、ここが壊れると「選んでいない札の置き先」が光る。
describe('americanToadSourceCard', () => {
  const tableau = emptyTableau();
  tableau[1] = [
    { card: card('SPADE', 9), faceUp: true },
    { card: card('HEART', 8), faceUp: true },
  ];

  it('returns nothing for no source', () => {
    expect(americanToadSourceCard(tableau, [], [], null)).toBeUndefined();
    expect(americanToadSourceCard(tableau, [], [], undefined)).toBeUndefined();
  });

  it('takes the top of the reserve and the waste', () => {
    expect(
      americanToadSourceCard(tableau, [card('CLOVER', 2), card('CLOVER', 3)], [], { zone: 'reserve' })?.value,
    ).toBe(3);
    expect(americanToadSourceCard(tableau, [], [card('DIAMOND', 4)], { zone: 'waste' })?.value).toBe(4);
  });

  it('returns nothing when the reserve or waste is empty', () => {
    expect(americanToadSourceCard(tableau, [], [], { zone: 'reserve' })).toBeUndefined();
    expect(americanToadSourceCard(tableau, [], [], { zone: 'waste' })).toBeUndefined();
  });

  it('takes the named card of a column, defaulting to its top', () => {
    expect(americanToadSourceCard(tableau, [], [], { zone: 'tableau', col: 1, cardIndex: 0 })?.value).toBe(9);
    expect(americanToadSourceCard(tableau, [], [], { zone: 'tableau', col: 1 })?.value).toBe(8);
  });

  it('returns nothing for a column, index or zone that names no card', () => {
    expect(americanToadSourceCard(tableau, [], [], { zone: 'tableau' })).toBeUndefined();
    expect(americanToadSourceCard(tableau, [], [], { zone: 'tableau', col: 0 })).toBeUndefined();
    expect(americanToadSourceCard(tableau, [], [], { zone: 'tableau', col: 99 })).toBeUndefined();
    expect(americanToadSourceCard(tableau, [], [], { zone: 'tableau', col: 1, cardIndex: 7 })).toBeUndefined();
    expect(americanToadSourceCard(tableau, [], [], { zone: 'foundation', col: 0 })).toBeUndefined();
  });
});
