import type { Card } from '../types/card';

/** Legal foundation and waste destinations for a selected Crazy Quilt card. */
export interface CrazyQuiltLegalTargets {
  /** Foundation indices that accept the selected card. */
  foundation: boolean[];
  /** Whether the selected quilt card may be played onto the waste. */
  waste: boolean;
}

/**
 * Calculates the selected card's Crazy Quilt destinations.
 *
 * Sync: `CrazyQuilt.canPlaceOnFoundation` and `crazyQuiltAdjacentRank` in
 * `internal/domain/CrazyQuilt.go` (the foundation suit/rank rule and waste
 * adjacency rule are mirrored here for destination highlighting only).
 */
export function crazyQuiltLegalTargets(
  foundation: Card[][],
  foundationAscending: boolean[] | undefined,
  wasteTop: Card | null,
  selectedCard: Card | null,
  selectedFromQuilt: boolean,
): CrazyQuiltLegalTargets {
  const foundationSuits = ['SPADE', 'CLOVER', 'HEART', 'DIAMOND', 'SPADE', 'CLOVER', 'HEART', 'DIAMOND'];
  const validFoundation = foundation.map((pile, idx) => {
    if (!selectedCard || pile.length === 0 || pile.length >= 13) return false;
    if (selectedCard.design !== foundationSuits[idx]) return false;
    const top = pile[pile.length - 1];
    return foundationAscending?.[idx] !== false
      ? selectedCard.value === top.value + 1
      : selectedCard.value === top.value - 1;
  });

  return {
    foundation: validFoundation,
    waste: !!selectedCard && selectedFromQuilt && !!wasteTop && Math.abs(selectedCard.value - wasteTop.value) === 1,
  };
}
