import type { Card } from '../types/card';
import { valueName } from './cardUtils';

const DESIGN_SYMBOLS: Record<string, string> = {
  SPADE: '♠',
  HEART: '♥',
  DIAMOND: '♦',
  CLOVER: '♣',
};

/** Return accessible alt text for a card: "ジョーカー" for jokers, "♠ A" style for normal cards. */
export function cardAlt(card: Card): string {
  if (card.design === 'JOKER') return 'ジョーカー';
  const symbol = DESIGN_SYMBOLS[card.design] ?? card.design;
  return `${symbol} ${valueName(card.value)}`;
}
