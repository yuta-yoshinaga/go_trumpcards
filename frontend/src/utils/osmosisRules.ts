import type { Card } from '../types/card';

/** The set of rank values (1–13) present in a foundation pile. */
function ranksIn(pile: Card[]): Set<number> {
  return new Set(pile.map((c) => c.value));
}

/**
 * Ranks that can currently be added to foundation row `i`, mirroring the Osmosis
 * domain rule (`Osmosis.canPlaceOnFoundation`): an empty row accepts only the base
 * rank (rows ≥1 also need the row above to be non-empty); the base row (0) with a
 * fixed suit accepts any not-yet-placed rank; row i≥1 accepts only ranks already
 * present in the row above and not yet in the row itself.
 */
export function osmosisAllowedRanks(foundation: Card[][], baseRank: number, i: number): number[] {
  const pile = foundation[i];
  if (pile.length === 0) {
    if (i >= 1 && foundation[i - 1].length === 0) return [];
    return [baseRank];
  }
  const have = ranksIn(pile);
  if (i === 0) {
    const all: number[] = [];
    for (let r = 1; r <= 13; r++) if (!have.has(r)) all.push(r);
    return all;
  }
  const above = ranksIn(foundation[i - 1]);
  return [...above].filter((r) => !have.has(r)).sort((a, b) => a - b);
}
