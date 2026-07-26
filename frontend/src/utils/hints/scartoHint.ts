import type { ScartoResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns a frontend {@link HintResult} for Scarto (スカルト), or null when no
 * suggestion is available.
 *
 * Like French Tarot, the hint is computed entirely by the Go backend and
 * surfaced on the response's `hint` field (with a `reason` i18n suffix such as
 * `scarto_weak`, `lead_low`, `follow_win`, `follow_duck`, or `play_excuse`).
 * This adapter re-maps that server hint into the frontend HintResult shape so
 * the shared {@link useGameHint} tooltip can render it. The `targetAction` is
 * fixed to `play` because every hint ultimately points the player at a decision.
 */
export function getScartoHint(state: ScartoResponse): HintResult | null {
  const hint = state.hint;
  if (!hint?.reason) return null;
  return {
    targetAction: 'play',
    reason: `hint.${hint.reason}`,
    confidence: 'moderate',
  };
}
