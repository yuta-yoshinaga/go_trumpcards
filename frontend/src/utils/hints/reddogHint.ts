import type { RedDogResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { RedDogPhase } from '../../types/phases';

/**
 * Threshold spread above which raising is positive expected value.
 * House-edge math: with a 10-card spread (4-11) the win probability is
 * (spread / 11). Raising adds equity once spread × multiplier > raise loss
 * probability. Spread ≥ 7 (multiplier 1:1) gives p(win) ≈ 0.64 — strong raise.
 */
const RAISE_THRESHOLD = 7;

/**
 * Returns a Red Dog hint for the spread decision phase.
 * Only suggests raise on large spreads (≥7) where the math favors it; otherwise stay.
 */
export function getReddogHint(state: RedDogResponse): HintResult | null {
  if (state.phase !== RedDogPhase.SPREAD_DECISION) return null;
  if (state.spread >= RAISE_THRESHOLD) {
    return { targetAction: 'raise', reason: 'hint.largeSpread', confidence: 'strong' };
  }
  return { targetAction: 'stay', reason: 'hint.smallSpread', confidence: 'moderate' };
}
