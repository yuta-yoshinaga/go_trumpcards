import type { SuecaResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns a frontend {@link HintResult} for Sueca, or null when no suggestion is
 * available.
 *
 * Like Tute and Tressette, Sueca's hint is computed entirely by the Go backend
 * and surfaced on the response's `hint` field (with a `reason` i18n suffix).
 * This adapter re-maps that server hint into the frontend HintResult shape so
 * the shared {@link useGameHint} tooltip can render it. The `targetAction` is
 * fixed to `play` because the only action is playing a card.
 */
export function getSuecaHint(state: SuecaResponse): HintResult | null {
  const hint = state.hint;
  if (!hint?.reason) return null;
  return {
    targetAction: 'play',
    reason: `hint.${hint.reason}`,
    confidence: 'moderate',
  };
}
