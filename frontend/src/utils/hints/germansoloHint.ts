import type { GermanSoloResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns a frontend {@link HintResult} for German Solo, or null when no
 * suggestion is available.
 *
 * German Solo's hint is computed entirely by the Go backend and surfaced on the
 * response's `hint` field (with a `reason` i18n suffix such as `lead_high`,
 * `lead_low`, `follow_win`, `follow_duck`, `give_partner`, `discard_low`,
 * `call_ace`, `bid_frage`, `bid_solo`, `bid_tout`, or `bid_pass`). Reading the
 * server's reason rather than re-deriving the thresholds here is what keeps the
 * advice the page shows identical to the one the CPU acts on. This adapter re-maps
 * that server hint into the frontend HintResult shape so the shared
 * {@link hooks/useGameHint.useGameHint | useGameHint} tooltip can render it. The `targetAction` is fixed to
 * `play` because every hint ultimately points the player at a card.
 */
export function getGermanSoloHint(state: GermanSoloResponse): HintResult | null {
  const hint = state.hint;
  if (!hint?.reason) return null;
  return {
    targetAction: 'play',
    reason: `hint.${hint.reason}`,
    confidence: 'moderate',
  };
}
