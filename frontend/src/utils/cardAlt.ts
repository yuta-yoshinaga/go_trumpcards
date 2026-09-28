import i18n from '../i18n';
import type { Card } from '../types/card';
import { valueName } from './cardUtils';

const DESIGN_SYMBOLS: Record<string, string> = {
  SPADE: '♠',
  HEART: '♥',
  DIAMOND: '♦',
  CLOVER: '♣',
};

const RED_DESIGNS = new Set(['HEART', 'DIAMOND']);

const SUIT_SYMBOLS_BY_INDEX = ['', '♠', '♣', '♥', '♦'] as const;

/**
 * Suit symbol for the numeric suit index the Go backend sends (1=♠ 2=♣ 3=♥ 4=♦).
 * Any other value (0 = no suit / undecided, or out of range) returns `fallback`,
 * so each caller keeps its own placeholder ("", "-", "?", "none").
 */
export function suitSymbolAt(suit: number, fallback = ''): string {
  return suit >= 1 && suit <= 4 ? SUIT_SYMBOLS_BY_INDEX[suit] : fallback;
}

/** Return the suit symbol for a card design (e.g. SPADE → ♠), or the raw design when unknown. */
export function suitSymbol(design: string): string {
  return DESIGN_SYMBOLS[design] ?? design;
}

/** True when a card design maps to a known suit symbol (SPADE/CLOVER/HEART/DIAMOND). */
export function isSuitDesign(design: string): boolean {
  return design in DESIGN_SYMBOLS;
}

/** True when a suit design renders in red (hearts / diamonds). */
export function isRedSuitDesign(design: string): boolean {
  return RED_DESIGNS.has(design);
}

/** Return accessible alt text for a card: procedural cards use their descriptor
 * label/glyph, "Joker" for jokers, "♠ A" style for normal cards. */
export function cardAlt(card: Card): string {
  if (card.label) return card.glyph ? `${card.label} ${card.glyph}` : card.label;
  if (card.design === 'JOKER') return i18n.t('common:card.joker');
  return `${suitSymbol(card.design)} ${valueName(card.value)}`;
}
