import type { KlaverjasResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns a frontend {@link HintResult} for Klaverjas, or null when no
 * suggestion is available.
 *
 * Like Sueca and Tute, Klaverjas's hint is computed entirely by the Go backend
 * and surfaced on the response's `hint` field (with a `reason` i18n suffix).
 * This adapter re-maps that server hint into the frontend HintResult shape so
 * the shared {@link hooks/useGameHint.useGameHint | useGameHint} tooltip can render it. The `targetAction` is
 * fixed to `play` because the only action is playing a card.
 */
export function getKlaverjasHint(state: KlaverjasResponse): HintResult | null {
  const hint = state.hint;
  if (!hint?.reason) return null;
  return {
    targetAction: 'play',
    reason: `hint.${hint.reason}`,
    confidence: 'moderate',
  };
}
