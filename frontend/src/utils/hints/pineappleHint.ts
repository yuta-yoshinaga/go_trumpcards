import type { PineappleResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { getHoldemBaseHint } from './holdemBaseHint';

/** Returns a frontend HintResult for Pineapple Poker, or null if no suggestion available. */
export function getPineappleHint(state: PineappleResponse): HintResult | null {
  if (state.isDiscardPhase) {
    const humanIdx = state.players.findIndex((p) => p.isHuman);
    if (humanIdx < 0) return null;
    if (state.discardDone[humanIdx]) return null;
    return { targetAction: 'discard', reason: 'hint.discardWeakest', confidence: 'moderate' };
  }

  return getHoldemBaseHint(state);
}
