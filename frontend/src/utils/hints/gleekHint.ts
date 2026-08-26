import type { GleekResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns a frontend {@link HintResult} for Gleek, or null when no suggestion
 * is available.
 *
 * Gleek's hint is computed entirely by the Go backend and surfaced on the
 * response's `hint` field, with a `reason` i18n suffix — `bid_raise`,
 * `bid_pass`, `discard_stock`, `lead_high`, `follow_win`, `follow_duck` or
 * `discard_honour`. Reading the server's reason rather than re-deriving the
 * thresholds here is what keeps the advice the page shows identical to the one
 * the CPU acts on. This adapter re-maps it into the frontend HintResult shape
 * so the shared {@link hooks/useGameHint.useGameHint | useGameHint} tooltip can
 * render it.
 */
export function getGleekHint(state: GleekResponse): HintResult | null {
  const hint = state.hint;
  if (!hint?.reason) return null;
  return {
    targetAction: 'play',
    reason: `hint.${hint.reason}`,
    confidence: 'moderate',
  };
}
