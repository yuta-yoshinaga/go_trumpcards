import type { Card } from '../types/card';
import { suitName } from './cardUtils';

/**
 * True when a card is one of Auction Forty-Fives' three fixed top trumps:
 * the five of trumps, the jack of trumps, and the ace of hearts.
 *
 * The ace of hearts is a trump **whatever the trump suit is**, which is the
 * part no suit symbol on screen can convey. Holding a top trump also exempts
 * the player from having to follow suit (reneging).
 *
 * Mirrors `isTopTrump` in `internal/domain/FortyFives.go`.
 *
 * @param card - The card to classify.
 * @param trumpSuit - The numeric trump suit (1..4 as in the Go card designs); `0` while no trump is set.
 * @returns Whether the card is a top trump.
 */
export function isFortyFivesTopTrump(card: Card, trumpSuit: number): boolean {
  if (card.design === 'HEART' && card.value === 1) return true;
  if (card.design !== suitName(trumpSuit)) return false;
  return card.value === 5 || card.value === 11;
}

/**
 * Positions of the top trumps within a hand, in hand order.
 *
 * @param cards - The hand to scan.
 * @param trumpSuit - The numeric trump suit (1..4); `0` while no trump is set.
 * @returns The indices of the top trumps.
 */
export function fortyFivesTopTrumpIndices(cards: readonly Card[], trumpSuit: number): number[] {
  const indices: number[] = [];
  cards.forEach((c, i) => {
    if (isFortyFivesTopTrump(c, trumpSuit)) indices.push(i);
  });
  return indices;
}
