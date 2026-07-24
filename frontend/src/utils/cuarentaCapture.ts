import type { Card } from '../types/card';

/**
 * Returns the set of table-card indices a given hand card would capture in
 * Cuarenta. Capture is a pure rank match: playing a card sweeps every table
 * card of the same rank (`value`) at once. This mirrors the Go domain's
 * `cuarentaRankMatchIndexes` exactly — Cuarenta has no sum/sequence capture,
 * so no other table cards are ever taken.
 *
 * @param handCard - the card the human is previewing, or `null` when none.
 * @param tableCards - the current face-up table cards.
 * @returns a set of indices into `tableCards` that would be captured.
 */
export function cuarentaCaptureIndices(handCard: Card | null, tableCards: readonly Card[]): Set<number> {
  const captured = new Set<number>();
  if (!handCard) return captured;
  for (let i = 0; i < tableCards.length; i++) {
    if (tableCards[i]?.value === handCard.value) {
      captured.add(i);
    }
  }
  return captured;
}
