import type { BasraResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns a frontend {@link HintResult} for Basra (Bastra), or null when no
 * suggestion is available.
 *
 * Basra's hint is computed entirely by the Go backend and surfaced on the
 * response's `hint` field (with a `reason` i18n suffix such as `basra_sweep`,
 * `jack_sweep`, `capture`, or `trail_low`). This adapter re-maps that server hint
 * into the frontend HintResult shape so the shared {@link useGameHint} tooltip can
 * render it. The `targetAction` is fixed to `play` because every hint points the
 * player at which card to play.
 */
export function getBasraHint(state: BasraResponse): HintResult | null {
  const hint = state.hint;
  if (!hint?.reason) return null;
  return {
    targetAction: 'play',
    reason: `hint.${hint.reason}`,
    confidence: 'moderate',
  };
}
