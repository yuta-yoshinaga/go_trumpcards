import type { KnockoutWhistResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns a frontend {@link HintResult} for Knockout Whist, or null when no
 * suggestion is available.
 *
 * Like Sedma and Mariáš, Knockout Whist's hint is computed entirely by the Go
 * backend and surfaced on the response's `hint` field (with a `reason` i18n
 * suffix such as `lead_high`, `follow_win`, `follow_duck`, or `discard_low`).
 * This adapter re-maps that server hint into the frontend HintResult shape so
 * the shared {@link hooks/useGameHint.useGameHint | useGameHint} tooltip can render it. The `targetAction` is
 * fixed to `play` because the only action is playing a card.
 */
export function getKnockoutWhistHint(state: KnockoutWhistResponse): HintResult | null {
  const hint = state.hint;
  if (!hint?.reason) return null;
  return {
    targetAction: 'play',
    reason: `hint.${hint.reason}`,
    confidence: 'moderate',
  };
}
