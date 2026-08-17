import type { AmericanToadTableauCard, Card } from '../types/card';

/** Where the selected card may legally go. */
export interface AmericanToadLegalTargets {
  /** Tableau column indices that accept the card. */
  tableau: Set<number>;
  /** Foundation indices that accept the card. */
  foundation: Set<number>;
}

/** Foundation index → suit, fixed so the layout does not move between deals. */
const FOUNDATION_SUITS = ['SPADE', 'CLOVER', 'HEART', 'DIAMOND', 'SPADE', 'CLOVER', 'HEART', 'DIAMOND'] as const;

/** A foundation is finished at thirteen cards. */
const FOUNDATION_TARGET = 13;

/** Ace follows King going up. Sync: `americanToadNextRank`. */
function nextRank(v: number): number {
  return v >= 13 ? 1 : v + 1;
}

/** King follows Ace going down. Sync: `americanToadPrevRank`. */
function prevRank(v: number): number {
  return v <= 1 ? 13 : v - 1;
}

/**
 * Legal destinations for the selected card.
 *
 * Sync: `AmericanToad.canPlaceOnTableau` / `canPlaceOnFoundation`.
 *
 * **タブローは同スートの降順、ファンデーションは同スートの昇順。**どちらも
 * A と K が地続き (A の下は K、K の上は A) で、ここを普通の 1..13 で書くと
 * 折り返しの手が置けない列として表示される。
 *
 * **空列はリザーブが残っている間は置き先にならない。**自動補充の対象なので、
 * 手で置くことはできない (#5559)。
 */
export function americanToadLegalTargets(
  tableau: readonly AmericanToadTableauCard[][],
  foundation: readonly Card[][],
  reserve: readonly Card[],
  baseRank: number,
  card: Card | null | undefined,
): AmericanToadLegalTargets {
  const result: AmericanToadLegalTargets = { tableau: new Set(), foundation: new Set() };
  if (!card) return result;

  tableau.forEach((col, idx) => {
    const top = col[col.length - 1]?.card;
    if (!top) {
      // 空列: リザーブが尽きて初めて手で置ける。
      if (reserve.length === 0) result.tableau.add(idx);
      return;
    }
    if (card.design === top.design && card.value === prevRank(top.value)) result.tableau.add(idx);
  });

  // baseRank が決まる前 (0) はどの基礎札も受け取らない。
  if (baseRank === 0) return result;

  foundation.forEach((pile, idx) => {
    if (FOUNDATION_SUITS[idx] !== card.design) return;
    if (pile.length === 0) {
      if (card.value === baseRank) result.foundation.add(idx);
      return;
    }
    if (pile.length >= FOUNDATION_TARGET) return;
    const top = pile[pile.length - 1];
    if (card.value === nextRank(top.value)) result.foundation.add(idx);
  });

  return result;
}
