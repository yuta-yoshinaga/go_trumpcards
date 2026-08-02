import type { FaroResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { FaroPhase } from '../../types/phases';

/**
 * Returns a frontend {@link HintResult} for Faro, or null when no suggestion is
 * available.
 *
 * Faro gives the frontend nothing to count from: the response carries only the
 * most recent turn's two cards, not the ranks already dealt, so any advice on
 * where to place a chip would be invented rather than derived. The hint stays
 * out of the betting layout entirely.
 *
 * The call is different, because it can be answered with arithmetic the player
 * can check: three cards, six possible orders, and the call pays 4:1. Guessing
 * loses money over time, so the hint says to skip it unless the player has been
 * keeping track — which is the whole historical point of a casekeeper.
 */
export function getFaroHint(state: FaroResponse): HintResult | null {
  if (state.gameEndFlag || state.phase !== FaroPhase.CALL) return null;

  // 呼べる札が 3 枚揃っていなければ確率の話にならない。
  return state.callCards.length === 3
    ? { targetAction: 'skipCall', reason: 'frontendHint.faroCallOdds', confidence: 'moderate' }
    : null;
}
