import type { ReversisResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns a frontend {@link HintResult} for Reversis, or null when no
 * suggestion is available.
 *
 * The marked cards cost chips as well as points, so refusing a trick that
 * holds one is close to forced; ordinary penalty cards and safe-trick discards
 * are judgement calls.
 */
export function getReversisHint(state: ReversisResponse): HintResult | null {
  const hint = state.hint;
  if (hint?.cardIndex === undefined) return null;

  return {
    targetAction: `card-${hint.cardIndex}`,
    reason: `hint.${hint.reason}`,
    confidence: hint.reason === 'reversisAvoidMarked' ? 'strong' : 'moderate',
  };
}
