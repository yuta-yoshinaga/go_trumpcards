/**
 * Number of cards taken when drawing from index `idx` of a Rummy 500 discard pile.
 *
 * Drawing from a chosen index takes that card plus every card above it (up to the
 * top of the pile), matching the domain rule `discardPile[idx:]` in
 * `internal/domain/Rummy500.go`. The pile is ordered oldest-first, so the top card
 * is the last element and picking from `idx` yields `pileLength - idx` cards.
 * Out-of-range indices return 0.
 */
export function rummy500PickupCount(pileLength: number, idx: number): number {
  if (idx < 0 || idx >= pileLength) return 0;
  return pileLength - idx;
}
