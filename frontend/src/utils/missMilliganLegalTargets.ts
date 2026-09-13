import type { Card } from '../types/common';
import type { MissMilliganTableauCard } from '../types/games/missmilligan';

/** The fixed foundation suit order used by Miss Milligan. */
const FOUNDATION_SUITS = ['SPADE', 'CLOVER', 'HEART', 'DIAMOND', 'SPADE', 'CLOVER', 'HEART', 'DIAMOND'] as const;

/** The legal destinations for a selected Miss Milligan card. */
export interface MissMilliganLegalTargets {
  /** Tableau column indices that accept the card. */
  tableau: Set<number>;
  /** Foundation indices that accept the card. */
  foundation: Set<number>;
}

/**
 * Returns the Miss Milligan destinations that accept `card`.
 *
 * domain_rule_quoted:
 * 	case fromZone == "tableau" && toZone == "foundation":
 * 		bc.writePresenterResponse(w, mi.MoveTableauToFoundation(*param.From.Col))
 *
 * 	card := fromCards[len(fromCards)-1].Card
 * 	fIdx := mm.findFoundation(card)
 *
 * edge_cases:
 * - Empty tableau columns accept Kings only.
 * - Empty foundations accept only the Ace of their fixed suit.
 * - A null card produces no targets.
 * - The source tableau column is excluded when it is provided.
 * - 組札へは最上段しか行かないので、途中の札を選んでいるときは組札の候補を出さない。
 */
export function missMilliganLegalTargets(
  tableau: readonly MissMilliganTableauCard[][],
  foundation: readonly Card[][],
  card: Card | null | undefined,
  sourceColumn?: number,
  sourceCardIndex?: number,
): MissMilliganLegalTargets {
  const result: MissMilliganLegalTargets = { tableau: new Set(), foundation: new Set() };
  if (!card) return result;
  const isRed = card.design === 'HEART' || card.design === 'DIAMOND';
  tableau.forEach((column, index) => {
    if (index === sourceColumn) return;
    const top = column[column.length - 1]?.card;
    if (!top) {
      if (card.value === 13) result.tableau.add(index);
      return;
    }
    const topIsRed = top.design === 'HEART' || top.design === 'DIAMOND';
    if (card.value === top.value - 1 && isRed !== topIsRed) result.tableau.add(index);
  });
  const canMoveTableauSelectionToFoundation =
    sourceColumn === undefined || sourceCardIndex === tableau[sourceColumn]?.length - 1;
  if (!canMoveTableauSelectionToFoundation) return result;
  foundation.forEach((pile, index) => {
    const top = pile[pile.length - 1];
    if (!top) {
      if (card.value === 1 && FOUNDATION_SUITS[index] === card.design) result.foundation.add(index);
    } else if (card.design === top.design && card.value === top.value + 1) {
      result.foundation.add(index);
    }
  });
  return result;
}
