import type { Card } from '../types/card';

/**
 * Pitch "Game" point pip values. Only A, K, Q, J, and 10 contribute; all other
 * cards score 0. The "Game" point at end-of-hand goes to whoever captured the
 * most pips.
 */
const PIP_VALUE: Readonly<Record<number, number>> = {
  1: 4, // Ace
  10: 10,
  11: 1, // Jack
  12: 2, // Queen
  13: 3, // King
};

/** Returns the Pitch Game-point pip value for a single card (0 for 2..9). */
export function pitchPipValue(value: number): number {
  return PIP_VALUE[value] ?? 0;
}

/**
 * Returns the total Pitch Game-point pip value across a hand. Used by the
 * UI assist badge (#1891) so players don't have to mentally sum their hand
 * every bid.
 */
export function pitchHandPips(hand: readonly Card[]): number {
  return hand.reduce((sum, c) => sum + pitchPipValue(c.value), 0);
}

/** Returns the per-card breakdown for the hand, in input order. */
export function pitchHandPipBreakdown(hand: readonly Card[]): { value: number; pips: number }[] {
  return hand.map((c) => ({ value: c.value, pips: pitchPipValue(c.value) }));
}
