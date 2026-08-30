import type { Card } from '../types/card';
import type { StreetsAndAlleysTableauCard } from '../types/games/streetsandalleys';

/** Where the selected card may legally go. */
export interface StreetsAndAlleysLegalTargets {
  /** Tableau column indices that accept the card. */
  tableau: Set<number>;
  /** Foundation indices that accept the card. */
  foundation: Set<number>;
}

/**
 * Legal destinations for the selected card.
 *
 * Sync: `StreetsAndAlleys.canPlaceOnTableau` / `canPlaceOnFoundationPile`
 * (`internal/domain/StreetsAndAlleys.go:371-383`).
 *
 * - The tableau ignores suit and colour entirely: only `value - 1` matters, and an empty
 *   column takes anything. Porting a sister game's alternating-colour condition would
 *   leave legal columns dark.
 * - Only one foundation is reported, not every accepting pile. `MoveTableauToFoundation`
 *   takes the *source* column and the domain's `findFoundation` scans for the first pile
 *   that accepts (`StreetsAndAlleys.go:386-393`), so with two empty foundations an ace is
 *   "legal" on both but always lands on the first. Ringing both would offer a choice the
 *   server does not have.
 */
export function streetsAndAlleysLegalTargets(
  tableau: readonly (readonly StreetsAndAlleysTableauCard[])[],
  foundation: readonly Card[][],
  card: Card | null | undefined,
): StreetsAndAlleysLegalTargets {
  const result: StreetsAndAlleysLegalTargets = { tableau: new Set(), foundation: new Set() };
  if (!card) return result;

  tableau.forEach((col, idx) => {
    if (col.length === 0) {
      result.tableau.add(idx);
      return;
    }
    const top = col[col.length - 1]?.card;
    if (top && card.value === top.value - 1) {
      result.tableau.add(idx);
    }
  });

  for (let idx = 0; idx < foundation.length; idx++) {
    const pile = foundation[idx];
    if (pile.length === 0) {
      if (card.value === 1) {
        result.foundation.add(idx);
        break;
      }
    } else {
      const top = pile[pile.length - 1];
      if (top && card.design === top.design && card.value === top.value + 1) {
        result.foundation.add(idx);
        break;
      }
    }
  }

  return result;
}
