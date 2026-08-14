import type { ShelemResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/**
 * Returns a frontend {@link HintResult} for Shelem, or null when no suggestion
 * is available.
 *
 * Neither the bidding hint nor the discard hint names a card: the first
 * recommends a points figure, the second a trump suit. Feeding a point card to
 * a partner who is already winning is close to automatic, since card points
 * are exactly what the contract is settled on.
 */
export function getShelemHint(state: ShelemResponse): HintResult | null {
  const hint = state.hint;
  if (!hint) return null;

  if (hint.cardIndex === undefined) {
    if (hint.reason === 'shelemBid') {
      return { targetAction: `bid-${hint.value}`, reason: `hint.${hint.reason}`, confidence: 'moderate' };
    }
    if (hint.reason === 'shelemDiscard') {
      return { targetAction: `trump-${hint.suit}`, reason: `hint.${hint.reason}`, confidence: 'moderate' };
    }
    return { targetAction: 'pass', reason: `hint.${hint.reason}`, confidence: 'moderate' };
  }

  return {
    targetAction: `card-${hint.cardIndex}`,
    reason: `hint.${hint.reason}`,
    confidence: hint.reason === 'shelemFeedPartner' ? 'strong' : 'moderate',
  };
}
