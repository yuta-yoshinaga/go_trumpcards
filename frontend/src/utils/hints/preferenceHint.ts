import type { PreferenceResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns a frontend {@link HintResult} for Préférence, or null when no
 * suggestion is available.
 *
 * Like Solo Whist, Préférence's hint is computed entirely by the Go backend and
 * surfaced on the response's `hint` field (with a `reason` i18n suffix). This
 * adapter re-maps that server hint into the frontend HintResult shape so the
 * shared {@link useGameHint} tooltip can render it. The `targetAction` is fixed
 * to `play` because the hint only applies during the trick-play phase.
 */
export function getPreferenceHint(state: PreferenceResponse): HintResult | null {
  const hint = state.hint;
  if (!hint?.reason) return null;
  return {
    targetAction: 'play',
    reason: `hint.${hint.reason}`,
    confidence: 'moderate',
  };
}
