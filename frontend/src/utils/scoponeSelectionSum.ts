import type { Card } from '../types/card';

/** A capture attempt in progress: what the selected table cards add up to, and what they must match. */
export interface ScoponeSelection {
  /** Total of the selected table cards. */
  sum: number;
  /** Value the total has to reach, taken from the chosen hand card. */
  target: number;
}

/**
 * The running capture total for Scopone, or null when no hand card is chosen.
 *
 * Unlike Escoba's fixed 15, the target moves with the hand card, so the player
 * is adding toward a different number every turn (#4767).
 * @param handCard - The chosen hand card, if any.
 * @param tableCards - All cards on the table.
 * @param tableIndices - Indices of the selected table cards.
 * @returns The sum and its target, or null.
 */
export function scoponeSelectionSum(
  handCard: Card | null | undefined,
  tableCards: readonly Card[],
  tableIndices: readonly number[],
): ScoponeSelection | null {
  if (!handCard) return null;
  const sum = tableIndices.reduce((total, idx) => total + (tableCards[idx]?.value ?? 0), 0);
  return { sum, target: handCard.value };
}
