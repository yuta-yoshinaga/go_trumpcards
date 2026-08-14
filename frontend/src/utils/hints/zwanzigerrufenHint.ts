import type { ZwanzigerrufenResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { ZwanzigerrufenPhase } from '../../types/phases';

/**
 * Returns a frontend {@link HintResult} for Zwanzigerrufen, or null when there
 * is nothing to advise.
 *
 * **Which card to play comes from the server** through `hint`; the page never
 * re-derives the follow-suit rules. What is added here is the shape of the
 * contract: under Trischaken taking points is what loses the deal, so the
 * advice is the opposite of every other contract.
 */
export function getZwanzigerrufenHint(state: ZwanzigerrufenResponse): HintResult | null {
  if (state.gameEndFlag) return null;

  if (state.phase === ZwanzigerrufenPhase.TRICK_END) {
    return { targetAction: 'next', reason: 'frontendHint.zwanzigerrufenNextTrick', confidence: 'strong' };
  }
  if (state.phase === ZwanzigerrufenPhase.ROUND_END) {
    return { targetAction: 'nextround', reason: 'frontendHint.zwanzigerrufenNextDeal', confidence: 'strong' };
  }
  if (!state.isHumanTurn) return null;

  if (state.phase === ZwanzigerrufenPhase.BID) {
    return { targetAction: 'bid', reason: 'frontendHint.zwanzigerrufenBidNeedsTrumps', confidence: 'moderate' };
  }
  if (state.phase === ZwanzigerrufenPhase.TALON) {
    return { targetAction: 'discard', reason: 'frontendHint.zwanzigerrufenBuryCheap', confidence: 'strong' };
  }
  if (state.contractName === 'trischaken') {
    return { targetAction: 'play', reason: 'frontendHint.zwanzigerrufenAvoidPoints', confidence: 'strong' };
  }
  return { targetAction: 'play', reason: 'frontendHint.zwanzigerrufenFollowSuit', confidence: 'moderate' };
}
