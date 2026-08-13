import type { DoubleAttackResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { DoubleAttackPhase } from '../../types/phases';

/** Up-card values the dealer busts from most often (with the tens removed). */
const WEAK_UP_CARD_MIN = 2;
const WEAK_UP_CARD_MAX = 6;

/**
 * Returns a frontend {@link HintResult} for Extra Bet Blackjack, or null when
 * there is nothing to advise.
 *
 * The advice that matters is the extra bet: raise on a weak up-card, decline on
 * a strong one. **With the tens removed the dealer busts less often**, so
 * raising into a strong up-card is worse here than the usual blackjack instinct
 * suggests.
 */
export function getDoubleattackHint(state: DoubleAttackResponse): HintResult | null {
  if (state.gameEndFlag) return null;

  if (state.phase === DoubleAttackPhase.ATTACK) {
    const up = state.dealerCards[0];
    if (!up) return null;
    const value = up.value === 1 ? 11 : Math.min(up.value, 10);
    if (value >= WEAK_UP_CARD_MIN && value <= WEAK_UP_CARD_MAX) {
      return { targetAction: 'attack', reason: 'frontendHint.doubleAttackWeakUp', confidence: 'moderate' };
    }
    return { targetAction: 'decline', reason: 'frontendHint.doubleAttackStrongUp', confidence: 'moderate' };
  }

  if (state.phase === DoubleAttackPhase.PLAY) {
    const hand = state.hands[state.activeHand];
    if (!hand) return null;
    if (hand.score <= 11) {
      return { targetAction: 'hit', reason: 'frontendHint.doubleAttackCannotBust', confidence: 'strong' };
    }
    if (hand.score >= 17) {
      return { targetAction: 'stand', reason: 'frontendHint.doubleAttackStandPat', confidence: 'strong' };
    }
    return { targetAction: 'hit', reason: 'frontendHint.doubleAttackBorderline', confidence: 'moderate' };
  }
  return null;
}
