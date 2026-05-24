import type { Card } from '../types/card';

/** Backend Mighty suit ints (matches CardDesignSpade…Diamond in Go domain). */
const SPADE = 1;
const CLOVER = 2;
const DIAMOND = 4;

/** Visual role a card plays in the current Mighty round. */
export type MightySpecialRole = 'mighty' | 'jokerCall' | 'joker' | 'partner';

/** Lookup card → role using the current `trumpSuit` and `partnerCard` from
 * the backend state. Returns `null` for ordinary cards.
 *
 * Mirrors the backend rules:
 *   - Mighty: ♦A when trumpSuit is ♠, otherwise ♠A.
 *   - JokerCall: ♠3 when trumpSuit is ♣, otherwise ♣3.
 *   - Joker: any card with design `JOKER`.
 *   - Partner: equals `partnerCard` by suit + value.
 */
export function mightySpecialRole(
  card: Card,
  trumpSuit: number,
  partnerCard: Card | null | undefined,
): MightySpecialRole | null {
  if (card.design === 'JOKER') return 'joker';
  if (isMighty(card, trumpSuit)) return 'mighty';
  if (isJokerCall(card, trumpSuit)) return 'jokerCall';
  if (partnerCard && card.design === partnerCard.design && card.value === partnerCard.value) return 'partner';
  return null;
}

function isMighty(card: Card, trumpSuit: number): boolean {
  if (card.value !== 1) return false;
  if (trumpSuit === SPADE) return card.design === 'DIAMOND';
  return card.design === 'SPADE';
}

function isJokerCall(card: Card, trumpSuit: number): boolean {
  if (card.value !== 3) return false;
  if (trumpSuit === CLOVER) return card.design === 'SPADE';
  return card.design === 'CLOVER';
}

/** Convert role into a short emoji glyph for the badge. */
export function mightyRoleGlyph(role: MightySpecialRole): string {
  switch (role) {
    case 'mighty':
      return '\u{1F451}'; // 👑
    case 'jokerCall':
      return '\u{1F3AF}'; // 🎯
    case 'joker':
      return '\u{1F0CF}'; // 🃏
    case 'partner':
      return '\u{1F91D}'; // 🤝
  }
}

// Re-export for tests / fixture authors that already work in backend suit ints.
export const MIGHTY_SUIT = {
  SPADE,
  CLOVER,
  HEART: 3,
  DIAMOND,
} as const;
