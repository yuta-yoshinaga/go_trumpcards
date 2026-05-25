import type { Card } from '../types/card';

/** Spite & Malice: K (13) acts as a wild that fills any foundation slot. */
const KING_VALUE = 13;
/** Spite & Malice: a foundation completes at Q (12) — the next card cannot stack on a full foundation. */
const FOUNDATION_TOP_COMPLETE = 12;

/**
 * Returns true when the goal pile's top card can legally be played onto at
 * least one foundation. K is wild (always playable on any non-complete
 * foundation); otherwise the card's value must equal `foundationTop + 1`.
 *
 * Used by the goal-pile affordance (#1886) so the UI nudges the player to
 * spend the goal pile — the actual win condition — whenever it is legal.
 */
export function isGoalTopPlayableToFoundation(top: Card | undefined, foundationTops: readonly number[]): boolean {
  if (!top) return false;
  if (top.value === KING_VALUE) {
    return foundationTops.some((t) => t < FOUNDATION_TOP_COMPLETE);
  }
  return foundationTops.some((t) => top.value === t + 1);
}
