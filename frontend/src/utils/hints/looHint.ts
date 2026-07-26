import type { LooResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns a frontend {@link HintResult} for Loo (Lanterloo), or null when no
 * suggestion is available.
 *
 * Like the other trick-takers, Loo's hint is computed entirely by the Go backend
 * and surfaced on the response's `hint` field (with a `reason` i18n suffix such
 * as `decide_play`, `decide_pass`, `lead_high`, `follow_win`, or `discard_low`).
 * This adapter re-maps that server hint into the frontend HintResult shape so the
 * shared {@link hooks/useGameHint.useGameHint | useGameHint} tooltip can render it. The `targetAction` is fixed to
 * `play` because every hint ultimately points the player at a decision.
 */
export function getLooHint(state: LooResponse): HintResult | null {
  const hint = state.hint;
  if (!hint?.reason) return null;
  return {
    targetAction: 'play',
    reason: `hint.${hint.reason}`,
    confidence: 'moderate',
  };
}
