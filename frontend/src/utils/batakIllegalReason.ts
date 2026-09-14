import type { Card } from '../types/common';
import type { BatakTrickCard } from '../types/games/batak';

/** Reasons why a Batak card is unavailable under the current play rules. */
export type BatakIllegalReason = 'spadesNotBroken' | 'followSuit' | 'mustTrumpSpade';

/**
 * Returns the reason a Batak card is unavailable, if any.
 *
 * This is the frontend counterpart to `internal/domain/Batak.go:530-560`
 * (`validatePlay`). If the rules change, update both implementations together.
 *
 * @param cardIndex - Index of the card being described.
 * @param cards - The human player's current hand.
 * @param currentTrick - Cards already played in the current trick.
 * @param spadesBroken - Whether spades have been broken.
 * @param validPlayIndices - Indices the backend says are legal to play.
 * @returns The rule identifier, or `undefined` when the card is playable.
 */
export function batakIllegalReason(
  cardIndex: number,
  cards: Card[],
  currentTrick: BatakTrickCard[],
  spadesBroken: boolean,
  validPlayIndices: number[],
): BatakIllegalReason | undefined {
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
