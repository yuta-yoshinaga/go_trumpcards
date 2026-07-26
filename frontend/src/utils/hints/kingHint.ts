import type { KingResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns a frontend {@link HintResult} for King, or null when no suggestion is
 * available.
 *
 * King's hint is computed entirely by the Go backend and surfaced on the
 * response's `hint` field (with a `reason` i18n suffix such as `avoid_low` or
 * `win_high`). This adapter re-maps that server hint into the frontend
 * HintResult shape so the shared {@link hooks/useGameHint.useGameHint | useGameHint} tooltip can render it. The
 * `targetAction` is fixed to `play` because every hint ultimately points the
 * player at a card.
 */
export function getKingHint(state: KingResponse): HintResult | null {
  const hint = state.hint;
  if (!hint?.reason) return null;
  return {
    targetAction: 'play',
    reason: `hint.${hint.reason}`,
    confidence: 'moderate',
  };
}
