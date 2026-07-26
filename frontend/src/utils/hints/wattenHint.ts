import type { WattenResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns a frontend {@link HintResult} for Watten (ヴァッテン), or null when no
 * suggestion is available.
 *
 * Watten's hint is computed entirely by the Go backend and surfaced on the
 * response's `hint` field, whose `action` is one of `declare`, `raise`, `play`,
 * `hold`, or `fold` and whose `reason` is an i18n suffix such as
 * `declare_strong`, `raise_strong`, `lead_trump`, `lead_plain`, `follow_win`,
 * `follow_dump`, `hold_ok`, or `fold_weak`. This adapter re-maps the server hint
 * into the frontend HintResult shape so the shared {@link useGameHint} tooltip
 * can render it; `targetAction` mirrors the server's action.
 */
export function getWattenHint(state: WattenResponse): HintResult | null {
  const hint = state.hint;
  if (!hint?.reason) return null;
  return {
    targetAction: hint.action || 'play',
    reason: `hint.${hint.reason}`,
    confidence: 'moderate',
  };
}
