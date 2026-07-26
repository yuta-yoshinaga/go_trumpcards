import type { PinochleResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Map server-side hint reason keys (sync: internal/domain/Pinochle.go) to
 * frontend-namespaced keys used by react-i18next.
 */
const REASON_MAP: Record<string, string> = {
  hint_bid: 'hintReason.hint_bid',
  hint_pass: 'hintReason.hint_pass',
  hint_trump: 'hintReason.hint_trump',
  hint_play: 'hintReason.hint_play',
};

/**
 * Returns a Pinochle hint by surfacing the server-computed recommendation.
 * The backend attaches a `hint` block whenever the player has an available
 * action; we map its raw reason key to the frontend namespace and mark it as a
 * strong suggestion since the backend already sees full state.
 */
export function getPinochleHint(state: PinochleResponse): HintResult | null {
  if (state.gameEndFlag) return null;
  const hint = state.hint;
  if (!hint?.reason) return null;
  return {
    targetAction: hint.cardIndex !== undefined ? 'play' : hint.pass ? 'pass' : 'bid',
    reason: REASON_MAP[hint.reason] ?? hint.reason,
    confidence: 'strong',
  };
}
