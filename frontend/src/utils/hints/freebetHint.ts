import type { FreeBetResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { FreeBetPhase } from '../../types/phases';

/** A hand at or above this total stands under basic strategy. */
const STAND_FROM = 17;
/** A hand at or below this total cannot bust on one card. */
const CANNOT_BUST_UP_TO = 11;
/** Up-card values the dealer busts from most often. */
const WEAK_UP_CARD_MIN = 2;
const WEAK_UP_CARD_MAX = 6;

/**
 * Returns a frontend {@link HintResult} for Free Bet Blackjack, or null when
 * there is nothing to advise.
 *
 * **The free raises come first, always.** The house pays the increment, so a
 * free double or free split cannot lose the player anything beyond the original
 * bet — taking one is correct even on a hand whose win rate is below even. That
 * is the whole reason the variant exists, and it is the one piece of advice a
 * player carrying ordinary blackjack instincts will get wrong.
 */
export function getFreebetHint(state: FreeBetResponse): HintResult | null {
  if (state.gameEndFlag) return null;
  if (state.phase !== FreeBetPhase.PLAY) return null;

  if (state.canFreeSplit) {
    return { targetAction: 'freesplit', reason: 'frontendHint.freeBetFreeIsFree', confidence: 'strong' };
  }
  if (state.canFreeDouble) {
    return { targetAction: 'freedouble', reason: 'frontendHint.freeBetFreeIsFree', confidence: 'strong' };
  }

  const hand = state.hands[state.activeHand];
  if (!hand) return null;
  if (hand.score <= CANNOT_BUST_UP_TO) {
    return { targetAction: 'hit', reason: 'frontendHint.freeBetCannotBust', confidence: 'strong' };
  }
  if (hand.score >= STAND_FROM) {
    return { targetAction: 'stand', reason: 'frontendHint.freeBetStandPat', confidence: 'strong' };
  }
  // **Between 12 and 16 the up-card decides it, exactly as the server's hint does.**
  // Returning one fixed answer here would contradict the hint the same board
  // produces server-side, and a player following both would be told two things.
  const up = state.dealerCards[0];
  const upValue = up ? (up.value === 1 ? 11 : Math.min(up.value, 10)) : 0;
  if (upValue >= WEAK_UP_CARD_MIN && upValue <= WEAK_UP_CARD_MAX) {
    return { targetAction: 'stand', reason: 'frontendHint.freeBetDealerMayBust', confidence: 'moderate' };
  }
  return { targetAction: 'hit', reason: 'frontendHint.freeBetChaseDealer', confidence: 'moderate' };
}
