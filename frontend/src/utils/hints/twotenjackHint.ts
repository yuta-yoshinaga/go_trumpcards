import type { TwoTenJackResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/** Map server-side hint reason keys (sync: internal/domain/TwoTenJack.go) to i18n keys. */
const REASON_MAP: Record<string, string> = {
  strategic_trump: 'hint.strategic_trump',
  lead: 'hint.lead',
  follow_suit: 'hint.follow_suit',
  trump_cut: 'hint.trump_cut',
  discard: 'hint.discard',
};

/**
 * Returns a Two Ten Jack hint by surfacing the server-computed recommendation.
 * The backend attaches a `hint` block to the response whenever the player has an
 * available action; we map its reason key to the frontend namespace and mark it
 * as strong since the backend already sees full state.
 */
export function getTwoTenJackHint(state: TwoTenJackResponse): HintResult | null {
  if (state.gameEndFlag) return null;
  const hint = state.hint;
  if (!hint?.reason) return null;
  return {
    targetAction: hint.cardIndex !== undefined ? 'play' : 'declare',
    reason: REASON_MAP[hint.reason] ?? hint.reason,
    confidence: 'strong',
  };
}
