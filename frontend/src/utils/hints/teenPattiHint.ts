import type { TeenPattiResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns a frontend {@link HintResult} for Teen Patti, or null when no
 * suggestion is available.
 *
 * Like Three Card Brag, Teen Patti's hint is computed entirely by the Go
 * backend and surfaced on the response's `hint` field (with a `reason` i18n
 * suffix). This adapter re-maps that server hint into the frontend HintResult
 * shape so the shared {@link useGameHint} tooltip can render it. The hint
 * carries an `action` string (`see`, `bet`, `raise`, `fold`, `show`, or
 * `sideshow`) used as the `targetAction`.
 */
export function getTeenPattiHint(state: TeenPattiResponse): HintResult | null {
  const hint = state.hint;
  if (!hint?.reason) return null;
  return {
    targetAction: hint.action || 'bet',
    reason: `hint.${hint.reason}`,
    confidence: 'moderate',
  };
}
