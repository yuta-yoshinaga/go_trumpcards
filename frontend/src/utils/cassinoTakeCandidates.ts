import type { Card } from '../types/card';

/**
 * Cards on the Cassino table that participate in at least one subset
 * whose values sum to the player's selected hand-card value.
 *
 * Returns:
 * - `indices`: a `Set<number>` of table-card indices touched by any matching subset.
 *   These should be highlighted in the UI as take candidates.
 *
 * Rules:
 * - Face cards (J=11, Q=12, K=13) cannot be combined; they only take by single
 *   matching face card. So when `targetValue >= 11`, only indices with the same
 *   exact value are returned.
 * - For number cards (1..10), all subsets summing to `targetValue` are valid.
 *   Single-card matches are included (a card equal to `targetValue`).
 * - Empty table or no matching subset returns an empty Set.
 *
 * The search enumerates subsets of up to `MAX_SUBSET_SIZE` table cards to keep
 * the worst case bounded; Cassino tables rarely exceed ~10 cards so a full
 * 2^N scan is also fine, but this is defense in depth.
 */
const MAX_SUBSET_SIZE = 12;

export function cassinoTakeCandidates(tableCards: readonly Card[], targetValue: number): { indices: Set<number> } {
  const indices = new Set<number>();
  if (tableCards.length === 0 || targetValue <= 0) return { indices };

  if (targetValue >= 11) {
    tableCards.forEach((c, i) => {
      if (c.value === targetValue) indices.add(i);
    });
    return { indices };
  }

  const n = Math.min(tableCards.length, MAX_SUBSET_SIZE);
  const total = 1 << n;
  for (let mask = 1; mask < total; mask += 1) {
    let sum = 0;
    for (let bit = 0; bit < n; bit += 1) {
      if ((mask >>> bit) & 1) {
        sum += tableCards[bit].value;
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
