import type { AmericanToadTableauCard, Card } from '../types/card';
import type { AmericanToadMoveZone } from '../types/games/americantoad';

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
 *
 * **空列は他のタブロー列からは決して埋められない** (`MoveTableauToTableau` の
 * 「空き列はリザーブと捨て札の出口であって、タブローの組み替えには使えない」、
 * #4417)。この規則は札そのものではなく*どこから来たか*で決まるので、
 * `fromZone` を渡さないと表現できない。
 *
 * @param fromZone - 動かす札の出どころ (`AmericanToadMoveZone.zone`)。省略すると空列の制限を掛けない。
 */
export function americanToadLegalTargets(
  tableau: readonly AmericanToadTableauCard[][],
  foundation: readonly Card[][],
  reserve: readonly Card[],
  baseRank: number,
  card: Card | null | undefined,
  fromZone?: string,
): AmericanToadLegalTargets {
  const result: AmericanToadLegalTargets = { tableau: new Set(), foundation: new Set() };
  if (!card) return result;

  tableau.forEach((col, idx) => {
    const top = col[col.length - 1]?.card;
    if (!top) {
      // 空列: リザーブが尽きて初めて手で置ける。ただしタブロー同士の
      // 組み替えでは埋められない。
      if (reserve.length === 0 && fromZone !== 'tableau') result.tableau.add(idx);
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

/**
 * The card a source zone refers to — the one whose destinations get rung.
 *
 * Lives beside {@link americanToadLegalTargets} because the two are always
 * called as a pair: this picks the card, that answers where it may go.
 *
 * A zone that names nothing (an empty pile, a column index past the end,
 * a run head that is no longer there) yields `undefined`, which the legality
 * pass reads as "highlight nothing" rather than as an error.
 *
 * @param source - The selected or hovered zone, or null.
 * @returns The card, or undefined when the zone names none.
 */
export function americanToadSourceCard(
  tableau: readonly AmericanToadTableauCard[][],
  reserve: readonly Card[],
  waste: readonly Card[],
  source: AmericanToadMoveZone | null | undefined,
): Card | undefined {
  if (!source) return undefined;
  if (source.zone === 'reserve') return reserve[reserve.length - 1];
  if (source.zone === 'waste') return waste[waste.length - 1];
  if (source.zone !== 'tableau' || source.col === undefined) return undefined;
  const pile = tableau[source.col] ?? [];
  // cardIndex 省略時は一番上の札 — ドラッグの掴み位置と同じ既定。
  const idx = source.cardIndex ?? pile.length - 1;
  return pile[idx]?.card ?? undefined;
}
