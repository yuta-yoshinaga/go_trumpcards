import type { QuadrilleResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns a frontend {@link HintResult} for Quadrille, or null when no
 * suggestion is available.
 *
 * Like Calabresella, Quadrille's hint is computed entirely by the Go backend and
 * surfaced on the response's `hint` field (with a `reason` i18n suffix such as
 * `lead_high`, `lead_low`, `follow_win`, `follow_duck`, `give_partner`,
 * `discard_low`, `bid_entrar`, `bid_solo`, or `bid_pass`). This adapter re-maps
 * that server hint into the frontend HintResult shape so the shared
 * {@link hooks/useGameHint.useGameHint | useGameHint} tooltip can render it. The `targetAction` is fixed to
 * `play` because every hint ultimately points the player at a card.
 */
export function getQuadrilleHint(state: QuadrilleResponse): HintResult | null {
  const hint = state.hint;
  if (!hint?.reason) return null;
  return {
    targetAction: 'play',
    reason: `hint.${hint.reason}`,
    confidence: 'moderate',
  };
}
