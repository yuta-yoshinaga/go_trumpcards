import trumpOrder from '../constants/fortyFivesTrumpOrder.json';

const DESIGN_SYMBOLS = {
  HEART: '♥',
  SPADE: '♠',
  CLOVER: '♣',
  DIAMOND: '♦',
} as const;

/** Formats the domain-derived trump order using the current trump suit. */
export function formatFortyFivesTrumpOrder(trumpSymbol: string): string[] {
  return trumpOrder
    .filter((card) => !(card.trump && card.value === 1 && trumpSymbol === '♥'))
    .map((card) => {
      const suit = card.trump ? trumpSymbol : DESIGN_SYMBOLS[card.design as keyof typeof DESIGN_SYMBOLS];
      const rank = { 1: 'A', 11: 'J', 12: 'Q', 13: 'K' }[card.value] ?? card.value;
      return `${suit}${rank}`;
    });
}
