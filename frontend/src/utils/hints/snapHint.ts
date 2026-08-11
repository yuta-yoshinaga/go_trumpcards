import type { SnapResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns a frontend {@link HintResult} for Snap, or null when no suggestion
 * is available.
 *
 * Calling when the match is showing is the one move with a right answer, so
 * it is the only one worth calling strong; turning a card and waiting are both
 * just what is left.
 */
export function getSnapHint(state: SnapResponse): HintResult | null {
  const hint = state.hint;
  if (!hint) return null;

  return {
    targetAction: hint.snap ? 'snap' : 'wait',
    reason: `hint.${hint.reason}`,
    confidence: hint.snap ? 'strong' : 'moderate',
  };
}
