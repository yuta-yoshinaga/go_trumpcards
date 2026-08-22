import type { SevenTwentySevenResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns a frontend {@link HintResult} for SevenTwentySeven, or null when there
 * is nothing to advise.
 *
 * The advice is computed by the Go backend and surfaced on `state.hint`. The
 * **reason names which side it is chasing** — "draw" on its own is not advice
 * when 7 and 27 pull in opposite directions, so the reason key carries the
 * whole point of the suggestion.
 */
export function getSevenTwentySevenHint(state: SevenTwentySevenResponse): HintResult | null {
  const hint = state.hint;
  if (!hint?.reason) return null;
  return {
    targetAction: hint.draw ? 'card' : 'stand',
    reason: `seventwentyseven:${hint.reason}`,
    confidence: 'moderate',
  };
}
