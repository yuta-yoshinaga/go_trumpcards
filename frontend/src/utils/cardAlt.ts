import i18n from '../i18n';
import type { Card } from '../types/card';
import { valueName } from './cardUtils';

const DESIGN_SYMBOLS: Record<string, string> = {
  SPADE: '♠',
  HEART: '♥',
  DIAMOND: '♦',
  CLOVER: '♣',
};

/** Return accessible alt text for a card: localized "Joker" for jokers, "♠ A" style for normal cards. */
export function cardAlt(card: Card): string {
  if (card.design === 'JOKER') return i18n.t('common:card.joker');
  const symbol = DESIGN_SYMBOLS[card.design] ?? card.design;
  return `${symbol} ${valueName(card.value)}`;
}
