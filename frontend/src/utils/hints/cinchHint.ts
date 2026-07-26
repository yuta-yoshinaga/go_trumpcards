import type { CinchResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns a frontend {@link HintResult} for Cinch (Double Pedro), or null when
 * no suggestion is available.
 *
 * Like the other bidding trick-takers, Cinch's hint is computed entirely by the
 * Go backend and surfaced on the response's `hint` field (with a `reason` i18n
 * suffix such as `bid_pass`, `bid_strong`, `name_trump`, `lead_strong`,
 * `trump_cut`, `follow_suit`, or `discard_low`). This adapter re-maps that server
 * hint into the frontend HintResult shape so the shared {@link useGameHint}
 * tooltip can render it. The `targetAction` is fixed to `play` because every
 * hint ultimately points the player at a decision.
 */
export function getCinchHint(state: CinchResponse): HintResult | null {
  const hint = state.hint;
  if (!hint?.reason) return null;
  return {
    targetAction: 'play',
    reason: `hint.${hint.reason}`,
    confidence: 'moderate',
  };
}
