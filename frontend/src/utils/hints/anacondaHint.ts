import type { AnacondaResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns a frontend {@link HintResult} for Anaconda, or null when no
 * suggestion is available.
 *
 * Anaconda's hint is computed entirely by the Go backend and surfaced on the
 * response's `hint` field. The hint carries an `action` (`pass` / `keep` /
 * `raise` / `call` / `fold`) mapped to the `targetAction` string, and a
 * `reason` suffix (`pass_weakest` / `keep_best` / `strong_hand` /
 * `medium_hand` / `weak_hand`) re-mapped into the frontend HintResult shape so
 * the shared {@link useGameHint} tooltip can render it. A `strong_hand`
 * suggestion reports strong confidence; everything else is moderate.
 */
export function getAnacondaHint(state: AnacondaResponse): HintResult | null {
  const hint = state.hint;
  if (!hint?.reason) return null;
  return {
    targetAction: hint.action,
    reason: `hint.${hint.reason}`,
    confidence: hint.reason === 'strong_hand' ? 'strong' : 'moderate',
  };
}
