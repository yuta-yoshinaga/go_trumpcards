import type { Card } from '../types/card';

/**
 * Scopa capture value of a card: A=1, 2-7 face value, J=8, Q=9, K=10.
 * Returns 0 for unknown values (defensive; should not occur for a 40/52-card deck).
 */
export function scopaCaptureValue(value: number): number {
  if (value >= 1 && value <= 10) return value;
  if (value === 11) return 8; // Jack
  if (value === 12) return 9; // Queen
  if (value === 13) return 10; // King
  return 0;
}

/**
 * Cards on the Scopa table that the player can capture by playing a card whose
 * capture value is `targetValue`.
 *
 * Returns:
 * - `indices`: a `Set<number>` of table-card indices that participate in a legal
 *   capture. These should be highlighted in the UI as take candidates.
 *
 * Rules (Scopa):
 * - If any single table card has a capture value equal to `targetValue`, the
 *   player MUST take exactly one such single card — so only those matching
 *   singles are returned (no multi-card subsets in that case).
 * - Otherwise, all subsets of table cards whose capture values sum to
 *   `targetValue` are legal; every index touched by a matching subset is returned.
 * - Empty table, non-positive target, or no matching subset returns an empty Set.
 *
 * The search enumerates subsets of up to `MAX_SUBSET_SIZE` table cards to keep
 * the worst case bounded; Scopa tables rarely exceed ~10 cards.
 */
const MAX_SUBSET_SIZE = 12;

export function scopaTakeCandidates(tableCards: readonly Card[], targetValue: number): { indices: Set<number> } {
  const indices = new Set<number>();
  if (tableCards.length === 0 || targetValue <= 0) return { indices };

  const values = tableCards.map((c) => scopaCaptureValue(c.value));

  // Forced-single rule: an exact single match overrides any subset capture.
  let hasSingle = false;
  values.forEach((v, i) => {
    if (v === targetValue) {
      indices.add(i);
      hasSingle = true;
    }
  });
  if (hasSingle) return { indices };

  const n = Math.min(values.length, MAX_SUBSET_SIZE);
  const total = 1 << n;
  for (let mask = 1; mask < total; mask += 1) {
    let sum = 0;
    for (let bit = 0; bit < n; bit += 1) {
      if ((mask >>> bit) & 1) {
        sum += values[bit];
        if (sum > targetValue) break;
      }
    }
    if (sum === targetValue) {
      for (let bit = 0; bit < n; bit += 1) {
        if ((mask >>> bit) & 1) indices.add(bit);
      }
    }
  }

  return { indices };
}
