import type { SlapjackResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/** Returns a Slapjack frontend hint or null. */
export function getSlapjackHint(state: SlapjackResponse): HintResult | null {
  if (state.gameEndFlag) return null;
  if (state.isTopJack) {
    return { targetAction: 'slap', reason: 'hint.slapJack', confidence: 'strong' };
  }
  if (state.isHumanTurn) {
    return { targetAction: 'step', reason: 'hint.flipCard', confidence: 'moderate' };
  }
  return null;
}
