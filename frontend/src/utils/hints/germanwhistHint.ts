import type { GermanWhistResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns a frontend {@link HintResult} for German Whist, or null when no
 * suggestion is available.
 *
 * The reason matters more here than in most games: in the first half the aim
 * inverts depending on whether the face-up card is worth having, so the same
 * "play a card" suggestion can mean *win this trick* or *lose it on purpose*.
 * Confidence is reported accordingly — ducking is a judgement call, taking a
 * card that is plainly better than what you hold is not.
 */
export function getGermanWhistHint(state: GermanWhistResponse): HintResult | null {
  const hint = state.hint;
  if (hint?.cardIndex === undefined) return null;

  return {
    targetAction: `card-${hint.cardIndex}`,
    reason: `hint.${hint.reason}`,
    confidence: hint.reason === 'germanWhistDuck' ? 'moderate' : 'strong',
  };
}
