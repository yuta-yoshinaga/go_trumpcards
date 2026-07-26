import type { TysiacResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns a frontend {@link HintResult} for Tysiąc (Thousand), or null when no
 * suggestion is available.
 *
 * Like Mariáš, Tysiąc's hint is computed entirely by the Go backend and
 * surfaced on the response's `hint` field (with a `reason` i18n suffix such as
 * `lead_low`, `lead_marriage`, `follow_win`, `follow_duck`, `discard_low`,
 * `bid_raise`, `bid_pass`, or `talon_discard`). This adapter re-maps that
 * server hint into the frontend HintResult shape so the shared
 * {@link useGameHint} tooltip can render it. The `targetAction` is fixed to
 * `play` because every hint ultimately points the player at a card.
 */
export function getTysiacHint(state: TysiacResponse): HintResult | null {
  const hint = state.hint;
  if (!hint?.reason) return null;
  return {
    targetAction: 'play',
    reason: `hint.${hint.reason}`,
    confidence: 'moderate',
  };
}
