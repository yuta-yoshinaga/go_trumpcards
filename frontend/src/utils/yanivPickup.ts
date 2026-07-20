/**
 * Yaniv discard-pickup rules.
 *
 * Mirrors the domain (`internal/domain/Yaniv.go`): after discarding, a player
 * may take a card from the previous player's discarded bunch, but only from its
 * two ends (`drawFromPickup` accepts `end` 0 = first or 1 = last). This holds
 * uniformly for singles, sets, and runs — there is no "any card of a set" rule.
 * For a single-card discard both ends collapse onto index 0.
 */

/**
 * Returns the indices of the previous discard that are legally pickup-able,
 * given the number of cards in that discard bunch.
 *
 * @param pickupCount - Number of cards in the previous discard (`pickupCards`).
 * @returns Sorted unique indices that may be taken (empty when nothing to take).
 */
export function pickupableIndices(pickupCount: number): number[] {
  if (pickupCount <= 0) return [];
  if (pickupCount === 1) return [0];
  return [0, pickupCount - 1];
}

/**
 * Reports whether the card at `index` of a `pickupCount`-card discard bunch is
 * legally pickup-able (i.e. is one of the two ends).
 *
 * @param index - Position within the previous discard bunch.
 * @param pickupCount - Number of cards in the previous discard bunch.
 * @returns True when the card at `index` may be taken.
 */
export function isPickupable(index: number, pickupCount: number): boolean {
  return pickupableIndices(pickupCount).includes(index);
}
