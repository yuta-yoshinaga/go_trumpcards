import type { SlobberhannesResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns a frontend {@link HintResult} for Slobberhannes, or null when no
 * suggestion is available.
 *
 * Confidence tracks how much is actually at stake: on a trick that carries a
 * penalty (first, last, or the one holding the Q of clubs) ducking is close to
 * forced, whereas on a safe trick the advice to shed a high card is a
 * judgement call.
 */
export function getSlobberhannesHint(state: SlobberhannesResponse): HintResult | null {
  const hint = state.hint;
  if (hint?.cardIndex === undefined) return null;

  return {
    targetAction: `card-${hint.cardIndex}`,
    reason: `hint.${hint.reason}`,
    confidence: hint.reason === 'slobberhannesAvoid' ? 'strong' : 'moderate',
  };
}
