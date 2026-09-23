import type { TappTarockResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { TappTarockPhase } from '../../types/phases';

/**
 * Returns a frontend {@link HintResult} for TappTarock, or null when there
 * is nothing to advise.
 *
 * **Which card to play comes from the server** through `hint`; the page never
 * re-derives the follow-suit rules. What is added here is the shape of the
 * contract: under Trischaken taking points is what loses the deal, so the
 * advice is the opposite of every other contract.
 */
export function getTappTarockHint(state: TappTarockResponse): HintResult | null {
  if (state.gameEndFlag) return null;

  if (state.phase === TappTarockPhase.TRICK_END) {
    return { targetAction: 'next', reason: 'frontendHint.tapptarockNextTrick', confidence: 'strong' };
  }
  if (state.phase === TappTarockPhase.ROUND_END) {
    return { targetAction: 'nextround', reason: 'frontendHint.tapptarockNextDeal', confidence: 'strong' };
  }
  if (!state.isHumanTurn) return null;

  if (state.phase === TappTarockPhase.BID) {
    return { targetAction: 'bid', reason: 'frontendHint.tapptarockBidNeedsTrumps', confidence: 'moderate' };
  }
  if (state.phase === TappTarockPhase.TALON) {
    return { targetAction: 'discard', reason: 'frontendHint.tapptarockBuryCheap', confidence: 'strong' };
  }
  if (state.contractName === 'trischaken') {
    return { targetAction: 'play', reason: 'frontendHint.tapptarockAvoidPoints', confidence: 'strong' };
  }
  return { targetAction: 'play', reason: 'frontendHint.tapptarockFollowSuit', confidence: 'moderate' };
}
