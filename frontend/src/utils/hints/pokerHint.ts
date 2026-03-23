import type { PokerResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { PokerPhase } from '../../types/phases';

/** Hand rank thresholds for hint logic. */
const RANK_ONE_PAIR = 1;
const RANK_TWO_PAIR = 2;
const RANK_THREE_OF_A_KIND = 3;

/** Returns a frontend HintResult for Poker, or null if no suggestion available. */
export function getPokerHint(state: PokerResponse): HintResult | null {
  const human = state.players.find((p) => p.isHuman);
  if (!human) return null;

  if (state.phase === PokerPhase.EXCHANGE) {
    return getExchangeHint(human.handRank);
  }

  if (state.phase === PokerPhase.DEAL || state.phase === PokerPhase.SECOND_BET) {
    return getBettingHint(human.handRank, human.folded);
  }

  return null;
}

/** Hint for the exchange phase based on current hand rank. */
function getExchangeHint(handRank: number): HintResult {
  if (handRank >= RANK_TWO_PAIR) {
    return { targetAction: 'stand', reason: 'hint.keepHand', confidence: 'strong' };
  }
  return { targetAction: 'exchange', reason: 'hint.exchangeWeak', confidence: 'moderate' };
}

/** Hint for betting phases based on hand rank. */
function getBettingHint(handRank: number, folded: boolean): HintResult | null {
  if (folded) return null;
  if (handRank >= RANK_THREE_OF_A_KIND) {
    return { targetAction: 'raise', reason: 'hint.strongHand', confidence: 'strong' };
  }
  if (handRank >= RANK_ONE_PAIR) {
    return { targetAction: 'call', reason: 'hint.decentHand', confidence: 'moderate' };
  }
  return { targetAction: 'fold', reason: 'hint.weakHand', confidence: 'moderate' };
}
