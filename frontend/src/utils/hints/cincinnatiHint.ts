import type { CincinnatiResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { CincinnatiPhase } from '../../types/phases';

/** Rank thresholds, matching the Go domain's advice. */
const TWO_PAIR = 3;
const TRIPS = 4;
const ONE_PAIR = 2;

/**
 * Returns a frontend {@link HintResult} for Cincinnati, or null when there is
 * nothing to advise.
 *
 * **The advice ordinary Hold'em instincts get wrong is how early a hand is
 * already made.** With five hole cards, a strong holding is often complete
 * before a single community card is turned — and there are five betting rounds
 * to pay for, so a hand that is not made yet is expensive to carry.
 *
 * The page cannot see the seat's rank before showdown (the server withholds
 * it), so this leans on the pot price the server does publish.
 */
export function getCincinnatiHint(state: CincinnatiResponse): HintResult | null {
  if (state.gameEndFlag) return null;
  if (state.phase !== CincinnatiPhase.BETTING || !state.isHumanTurn) return null;

  const seat = state.seats[state.humanSeat];
  if (!seat) return null;

  // handRank is only populated at showdown, so treat 0 as "not known yet".
  const rank = seat.handRank;

  if (state.toCall <= 0) {
    if (rank >= TWO_PAIR && state.canRaise) {
      return { targetAction: 'bet', reason: 'frontendHint.cincinnatiStrongEnoughToBet', confidence: 'moderate' };
    }
    return { targetAction: 'check', reason: 'frontendHint.cincinnatiSeeAnotherCard', confidence: 'strong' };
  }
  if (rank >= TRIPS && state.canRaise) {
    return { targetAction: 'raise', reason: 'frontendHint.cincinnatiStrongEnoughToRaise', confidence: 'moderate' };
  }
  if (rank >= ONE_PAIR) {
    return { targetAction: 'call', reason: 'frontendHint.cincinnatiWorthACall', confidence: 'moderate' };
  }
  // **五回ぶんのベットが残っている。** 役の無い手を安易に持ち越さない。
  const ante = state.config?.ante ?? 0;
  if (state.toCall <= ante) {
    return { targetAction: 'call', reason: 'frontendHint.cincinnatiCheapToStay', confidence: 'moderate' };
  }
  return { targetAction: 'fold', reason: 'frontendHint.cincinnatiNotWorthIt', confidence: 'moderate' };
}
