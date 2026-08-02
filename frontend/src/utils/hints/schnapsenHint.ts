import type { SchnapsenResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns a frontend {@link HintResult} for Schnapsen, or null when no suggestion is
 * available.
 *
 * The hint is computed by the Go backend and, since #4483, arrives on every
 * response rather than only the `hint` command's. This was a stub returning
 * null on the grounds that the reasoning lived server-side — true, and no
 * longer a reason to discard the answer once the server started sending it.
 *
 * The shapes are mutually exclusive, one per decision point, and each is tested
 * with `!== undefined` because zero is legal for all of them.
 */
export function getSchnapsenHint(state: SchnapsenResponse): HintResult | null {
  const hint = state.hint;
  if (hint?.cardIndex === undefined) return null;

  // **マリッジ宣言つきの手は別物。**同じ札でも申告するかどうかで点が変わる。
  return {
    targetAction: hint.isMarriage ? 'marriage' : `card-${hint.cardIndex}`,
    reason: `hint.${hint.reason}`,
    confidence: 'moderate',
  };
}
