import type { BaseballPokerResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { BaseballPhase } from '../../types/phases';

/** Rank thresholds, matching the Go domain's advice. */
const TWO_PAIR = 3;
const TRIPS = 4;
const FLUSH = 6;

/**
 * Returns a frontend {@link HintResult} for Baseball Poker, or null when there
 * is nothing to advise.
 *
 * **The instinct this game punishes is a wild-free sense of what a hand is
 * worth.** Eight wild cards in the deck raise the going rate, so two pair is
 * not the hand it is elsewhere — the thresholds here sit deliberately higher
 * than an ordinary stud game's.
 *
 * The page cannot see the seat's rank before showdown (the server withholds
 * it), so this leans on the pot price the server does publish.
 */
export function getBaseballpokerHint(state: BaseballPokerResponse): HintResult | null {
  if (state.gameEndFlag) return null;

  const seat = state.seats[state.humanSeat];
  if (!seat) return null;

  // **買い増しの返事が最優先。** その場で払うか降りるかしかない。
  if (state.isBuying || state.phase === BaseballPhase.BUY_IN) {
    if (!state.isBuying) return null;
    if (seat.handRank >= TWO_PAIR) {
      return { targetAction: 'pay', reason: 'frontendHint.baseballHandIsWorthTheBuy', confidence: 'moderate' };
    }
    if (state.buyCost * 3 <= seat.chips) {
      return { targetAction: 'pay', reason: 'frontendHint.baseballBuyIsCheapEnough', confidence: 'moderate' };
    }
    return { targetAction: 'fold', reason: 'frontendHint.baseballBuyCostsTooMuch', confidence: 'moderate' };
  }

  if (state.phase !== BaseballPhase.BETTING || !state.isHumanTurn) return null;

  // handRank is only populated at showdown, so treat 0 as "not known yet".
  const rank = seat.handRank;

  if (state.toCall <= 0) {
    if (rank >= TRIPS && state.canRaise) {
      return { targetAction: 'bet', reason: 'frontendHint.baseballStrongEnoughToBet', confidence: 'moderate' };
    }
    return { targetAction: 'check', reason: 'frontendHint.baseballSeeAnotherCard', confidence: 'strong' };
  }
  if (rank >= FLUSH && state.canRaise) {
    return { targetAction: 'raise', reason: 'frontendHint.baseballStrongEnoughToRaise', confidence: 'moderate' };
  }
  if (rank >= TWO_PAIR) {
    return { targetAction: 'call', reason: 'frontendHint.baseballWorthACall', confidence: 'moderate' };
  }
  const ante = state.config?.ante ?? 0;
  if (state.toCall <= ante) {
    return { targetAction: 'call', reason: 'frontendHint.baseballCheapToStay', confidence: 'moderate' };
  }
  return { targetAction: 'fold', reason: 'frontendHint.baseballWildsRaiseTheBar', confidence: 'moderate' };
}

/** True when the card value is wild for this game state. */
export function isBaseballWild(state: BaseballPokerResponse, value: number): boolean {
  return state.wildValues.includes(value);
}
