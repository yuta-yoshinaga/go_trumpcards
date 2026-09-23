import type { Card } from '../types/common';

/** Legal destinations for a Congress source card. */
export interface CongressLegalTargets {
  /** Tableau column indices that accept the source. */
  tableau: Set<number>;
  /** Foundation indices that accept the source. */
  foundation: Set<number>;
}

const FOUNDATION_SUITS = ['SPADE', 'CLOVER', 'HEART', 'DIAMOND', 'SPADE', 'CLOVER', 'HEART', 'DIAMOND'] as const;
const FOUNDATION_TARGET = 13;

/**
 * Returns the Congress destinations that can accept a selected source.
 *
 * domain_rule_quoted:
 * ```go
 * if len(c.tableau[pile]) == 0 { return true }     // 空き山は誰でも置ける…が下記参照
 * top := c.tableau[pile][len(c.tableau[pile])-1]
 * if top == nil { return false }
 * // 折り返しは無い。A の上には何も置けず、その山はそこで止まる。
 * return card.GetValue() == top.GetValue()-1
 *
 * if congressSuitOrder[fIdx] != card.GetDesign() { return false }
 * if len(pile) == 0 { return card.GetValue() == 1 }     // A から
 * if len(pile) >= CongressFoundationTarget { return false }
 * ... (この続きは自分で開いて読むこと)
 * ```
 *
 * The quoted foundation rule continues in the domain as:
 *
 * ```go
 * pile := c.foundation[fIdx]
 * if len(pile) == 0 {
 * 	return card.GetValue() == 1
 * }
 * if len(pile) >= CongressFoundationTarget {
 * 	return false
 * }
 * return card.GetValue() == pile[len(pile)-1].GetValue()+1
 * ```
 *
 * An empty tableau is available only to stock and waste sources. This mirrors
 * the page-level restriction used when dispatching a tableau move. Stock has
 * no card in the response, but its direct move is defined only for empty piles.
 *
 * @param tableau - Current tableau piles.
 * @param foundation - Current foundation piles.
 * @param card - Selected card, or null for the stock source.
 * @param sourceZone - Zone the selected card comes from.
 * @returns Sets of legal tableau and foundation indices.
 */
export function congressLegalTargets(
  tableau: readonly Card[][],
  foundation: readonly Card[][],
  card: Card | null | undefined,
  sourceZone: string | null | undefined,
): CongressLegalTargets {
  const result: CongressLegalTargets = { tableau: new Set(), foundation: new Set() };

  if (sourceZone === 'stock') {
    tableau.forEach((pile, idx) => {
      if (pile.length === 0) result.tableau.add(idx);
    });
    return result;
  }
  if (!card) return result;

  tableau.forEach((pile, idx) => {
    if (pile.length === 0) {
      if (sourceZone !== 'tableau') result.tableau.add(idx);
      return;
    }
    const top = pile[pile.length - 1];
    if (top && card.value === top.value - 1) result.tableau.add(idx);
  });

  foundation.forEach((pile, idx) => {
    if (FOUNDATION_SUITS[idx] !== card.design || pile.length >= FOUNDATION_TARGET) return;
    if (pile.length === 0) {
      if (card.value === 1) result.foundation.add(idx);
      return;
    }
    const top = pile[pile.length - 1];
    if (card.value === top.value + 1) result.foundation.add(idx);
  });

  return result;
}
