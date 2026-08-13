import type { BanLuckResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { BanLuckPhase } from '../../types/phases';

/** A hand at or above this total stands under basic strategy. */
const STAND_FROM = 17;
/** A hand at or below this total cannot bust on one card. */
const CANNOT_BUST_UP_TO = 11;
/** With four cards at or below this total, one more makes a Five Dragon. */
const DRAGON_CHASE_UP_TO = 15;
/** Five cards is the limit. */
const MAX_HAND_CARDS = 5;

/**
 * Returns a frontend {@link HintResult} for Ban Luck, or null when there is
 * nothing to advise.
 *
 * **The banker's obligation comes before any strategy.** Below 15 the banker
 * cannot stand at all, so advising anything else would point at a button the
 * page will not let you press.
 *
 * The other advice ordinary blackjack instincts get wrong is the Five Dragon:
 * with four cards at 21 or less, drawing beats standing even on 17, because
 * five cards under 21 beats an ordinary hand at any total.
 */
export function getBanluckHint(state: BanLuckResponse): HintResult | null {
  if (state.gameEndFlag) return null;
  if (state.phase !== BanLuckPhase.PLAY || !state.isHumanTurn) return null;

  if (state.mustHit) {
    return { targetAction: 'hit', reason: 'frontendHint.banLuckBankerMustHit', confidence: 'strong' };
  }

  const seat = state.seats[state.turnSeat];
  if (!seat) return null;

  if (seat.cards.length >= MAX_HAND_CARDS) {
    return { targetAction: 'stand', reason: 'frontendHint.banLuckHandFull', confidence: 'strong' };
  }
  if (seat.cards.length === MAX_HAND_CARDS - 1 && seat.score <= DRAGON_CHASE_UP_TO) {
    return { targetAction: 'hit', reason: 'frontendHint.banLuckChaseFiveDragon', confidence: 'strong' };
  }
  if (seat.score <= CANNOT_BUST_UP_TO) {
    return { targetAction: 'hit', reason: 'frontendHint.banLuckCannotBust', confidence: 'strong' };
  }
  if (seat.score >= STAND_FROM) {
    return { targetAction: 'stand', reason: 'frontendHint.banLuckStandPat', confidence: 'strong' };
  }
  return { targetAction: 'hit', reason: 'frontendHint.banLuckChaseBanker', confidence: 'moderate' };
}
