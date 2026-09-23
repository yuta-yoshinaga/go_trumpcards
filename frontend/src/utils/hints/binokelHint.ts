import type { BinokelResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { BinokelPhase } from '../../types/phases';

/**
 * Map server-side hint reason keys (sync: internal/domain/Binokel.go) to
 * frontend-namespaced keys used by react-i18next.
 */
const REASON_MAP: Record<string, string> = {
  hint_bid: 'hintReason.hint_bid',
  hint_pass: 'hintReason.hint_pass',
  hint_trump: 'hintReason.hint_trump',
  hint_play: 'hintReason.hint_play',
  hint_dabb: 'hintReason.hint_dabb',
};

/**
 * Returns a Binokel hint by surfacing the server-computed recommendation.
 * The backend attaches a `hint` block whenever the player has an available
 * action; we map its raw reason key to the frontend namespace and mark it as a
 * strong suggestion since the backend already sees full state.
 */
export function getBinokelHint(state: BinokelResponse): HintResult | null {
  if (state.gameEndFlag) return null;
  const hint = state.hint;
  if (!hint?.reason) return null;

  let targetAction = 'play';
  if (state.phase === BinokelPhase.DABB || (hint as { discardIndices?: number[] }).discardIndices) {
    targetAction = 'discard';
  } else if (hint.cardIndex !== undefined) {
    targetAction = 'play';
  } else if (hint.pass || hint.bidAmount === 0) {
    targetAction = 'pass';
  } else if (hint.suit !== undefined) {
    targetAction = 'trump';
  } else if (hint.bidAmount !== undefined && hint.bidAmount > 0) {
    targetAction = 'bid';
  }

  return {
    targetAction,
    reason: REASON_MAP[hint.reason] ?? hint.reason,
    confidence: 'strong',
  };
}
