/** Return a label for a card: "JOKER" for jokers, otherwise "DESIGN VALUE". */
export function cardLabel(card: { design: string; value: number }): string {
  if (card.design === 'JOKER') return 'JOKER';
  return `${card.design} ${card.value}`;
}

/** Return display name for a card value (1→'A', 11→'J', 12→'Q', 13→'K', else string). */
export function valueName(v: number): string {
  if (v === 1) return 'A';
  if (v === 11) return 'J';
  if (v === 12) return 'Q';
  if (v === 13) return 'K';
  return String(v);
}

const SUIT_NAMES: Record<number, string> = { 1: 'SPADE', 2: 'CLOVER', 3: 'HEART', 4: 'DIAMOND' };

/** Return suit name for a numeric suit index (1→'SPADE', 2→'CLOVER', 3→'HEART', 4→'DIAMOND'). */
export function suitName(suit: number): string {
  return SUIT_NAMES[suit] ?? '';
}
