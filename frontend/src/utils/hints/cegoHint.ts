import type { CegoResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns a frontend {@link HintResult} for Cego (チェゴ), or null when no
 * suggestion is available.
 *
 * Like Königrufen, the hint is computed entirely by the Go backend and surfaced
 * on the response's `hint` field (with a `reason` i18n suffix such as
 * `bid_take`, `bid_pass`, `contract_cego`, `contract_handspiel`, `keep_best`,
 * `lead_high`, `lead_low`, `follow_win`, or `follow_duck`). This adapter re-maps
 * that server hint into the frontend HintResult shape so the shared
 * {@link hooks/useGameHint.useGameHint | useGameHint} tooltip can render it. The `targetAction` is fixed to
 * `play` because every hint ultimately points the player at a decision.
 */
export function getCegoHint(state: CegoResponse): HintResult | null {
  const hint = state.hint;
  if (!hint?.reason) return null;
  return {
    targetAction: 'play',
    reason: `hint.${hint.reason}`,
    confidence: 'moderate',
  };
}
