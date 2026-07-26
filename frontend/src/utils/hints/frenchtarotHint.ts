import type { FrenchTarotResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns a frontend {@link HintResult} for French Tarot (フレンチタロット), or
 * null when no suggestion is available.
 *
 * Like Ulti, the hint is computed entirely by the Go backend and surfaced on the
 * response's `hint` field (with a `reason` i18n suffix such as `bid_take`,
 * `bid_pass`, `discard_weak`, `lead_high`, `lead_low`, `follow_win`,
 * `follow_duck`, or `play_excuse`). This adapter re-maps that server hint into
 * the frontend HintResult shape so the shared {@link useGameHint} tooltip can
 * render it. The `targetAction` is fixed to `play` because every hint ultimately
 * points the player at a decision.
 */
export function getFrenchTarotHint(state: FrenchTarotResponse): HintResult | null {
  const hint = state.hint;
  if (!hint?.reason) return null;
  return {
    targetAction: 'play',
    reason: `hint.${hint.reason}`,
    confidence: 'moderate',
  };
}
