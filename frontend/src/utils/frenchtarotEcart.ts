import type { Card } from '../types/card';

/**
 * Reason a French Tarot card may not be buried in the chien during the écart.
 *
 * - `king` — a suit King (Roi, value 14); never buriable.
 * - `excuse` — the Excuse (Fool); never buriable.
 * - `bout` — a bout trump (Petit = trump 1, or the 21); never buriable.
 * - `trump` — an ordinary trump (atout); only buriable when unavoidable.
 */
export type FrenchTarotUnburiableReason = 'king' | 'excuse' | 'bout' | 'trump';

/**
 * Classifies why a card cannot (normally) be buried into the chien during the
 * French Tarot écart, or returns `null` for a freely buriable card.
 *
 * Mirrors the backend rule (domain `validateDiscards` / `frenchTarotDiscardable`):
 * the Excuse and Kings are never buriable, the bouts (Petit = trump 1 and the
 * 21) are never buriable, and ordinary trumps may only be buried when the player
 * holds fewer than six freely buriable cards. Detection is color-based on the
 * serialized tarot face — the Excuse is gold and trumps (atouts) are purple,
 * while a King is a non-tarot suit card of value 14.
 */
export function frenchTarotUnburiableReason(card: Card): FrenchTarotUnburiableReason | null {
  if (card.color === 'gold') return 'excuse';
  if (card.color === 'purple') {
    return card.value === 1 || card.value === 21 ? 'bout' : 'trump';
  }
  if (card.value === 14) return 'king';
  return null;
}
