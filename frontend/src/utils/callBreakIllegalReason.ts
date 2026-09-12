import type { Card } from '../types/common';
import type { CallBreakTrickCard } from '../types/games/callbreak';

/** Reasons why a Call Break card is unavailable under the current play rules. */
export type CallBreakIllegalReason = 'spadesNotBroken' | 'followSuit' | 'mustTrumpSpade';

/**
 * Returns the reason a Call Break card is unavailable, if any.
 *
 * This is the frontend counterpart to `internal/domain/CallBreak.go:448-475`
 * (`validatePlay`). If the rules change, update both implementations together.
 *
 * @param cardIndex - Index of the card being described.
 * @param cards - The human player's current hand.
 * @param currentTrick - Cards already played in the current trick.
 * @param spadesBroken - Whether spades have been broken.
 * @param validPlayIndices - Indices the backend says are legal to play.
 * @returns The rule identifier, or `undefined` when the card is playable.
 */
export function callBreakIllegalReason(
  cardIndex: number,
  cards: Card[],
  currentTrick: CallBreakTrickCard[],
  spadesBroken: boolean,
  validPlayIndices: number[],
): CallBreakIllegalReason | undefined {
  if (validPlayIndices.includes(cardIndex)) return undefined;

  const card = cards[cardIndex];
  if (!card) return undefined;

  if (currentTrick.length === 0) {
    if (!spadesBroken && card.design === 'SPADE' && cards.some((handCard) => handCard.design !== 'SPADE')) {
      return 'spadesNotBroken';
    }
    return undefined;
  }

  const leadSuit = currentTrick[0].card.design;
  if (cards.some((handCard) => handCard.design === leadSuit)) {
    return card.design === leadSuit ? undefined : 'followSuit';
  }

  if (leadSuit !== 'SPADE' && cards.some((handCard) => handCard.design === 'SPADE')) {
    return card.design === 'SPADE' ? undefined : 'mustTrumpSpade';
  }

  return undefined;
}
