import type { Card } from '../types/card';

/** The set of rank values (1–13) present in a foundation pile. */
function ranksIn(pile: Card[]): Set<number> {
  return new Set(pile.map((c) => c.value));
}

/** True if suit `s` is already the assigned suit of a foundation row other than `except`. */
function suitAssigned(foundation: Card[][], s: string, except: number): boolean {
  return foundation.some((pile, i) => i !== except && pile.length > 0 && pile[0]?.design === s);
}

/**
 * Ranks that can currently be added to foundation row `i`, mirroring the Osmosis
 * domain rule (`Osmosis.canPlaceOnFoundation`): an empty row accepts only the base
 * rank (rows ≥1 also need the row above to be non-empty); the base row (0) with a
 * fixed suit accepts any not-yet-placed rank; row i≥1 accepts only ranks already
 * present in the row above and not yet in the row itself.
 */
export function osmosisAllowedRanks(foundation: Card[][], baseRank: number, i: number): number[] {
  const pile = foundation[i] ?? [];
  if (pile.length === 0) {
    if (i >= 1 && (foundation[i - 1]?.length ?? 0) === 0) return [];
    return [baseRank];
  }
  const have = ranksIn(pile);
  if (i === 0) {
    const all: number[] = [];
    for (let r = 1; r <= 13; r++) if (!have.has(r)) all.push(r);
    return all;
  }
  const above = ranksIn(foundation[i - 1] ?? []);
  return [...above].filter((r) => !have.has(r)).sort((a, b) => a - b);
}

/** Mirrors `Osmosis.canPlaceOnFoundation` for a specific candidate card. */
export function osmosisCanPlace(foundation: Card[][], baseRank: number, i: number, card: Card): boolean {
  const pile = foundation[i] ?? [];
  if (pile.length === 0) {
    if (card.value !== baseRank) return false;
    if (suitAssigned(foundation, card.design, i)) return false;
    if (i === 0) return true;
    return (foundation[i - 1]?.length ?? 0) > 0;
  }
  if (pile[0]?.design !== card.design) return false;
  if (i === 0) return true;
  return ranksIn(foundation[i - 1] ?? []).has(card.value);
}
