import i18n from '../i18n';
import type { Card } from '../types/card';
import { valueName } from './cardUtils';

const DESIGN_SYMBOLS: Record<string, string> = {
  SPADE: '♠',
  HEART: '♥',
  DIAMOND: '♦',
  CLOVER: '♣',
};

/** Return the suit symbol for a card design (e.g. SPADE → ♠), or the raw design when unknown. */
export function suitSymbol(design: Card['design']): string {
  return DESIGN_SYMBOLS[design] ?? design;
}

/** Return accessible alt text for a card: procedural cards use their descriptor
 * label/glyph, "Joker" for jokers, "♠ A" style for normal cards. */
export function cardAlt(card: Card): string {
  if (card.label) return card.glyph ? `${card.label} ${card.glyph}` : card.label;
  if (card.design === 'JOKER') return i18n.t('common:card.joker');
  return `${suitSymbol(card.design)} ${valueName(card.value)}`;
}
