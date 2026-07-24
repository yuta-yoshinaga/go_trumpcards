import type { Card } from '../types/card';

/**
 * Reason a Scarto card may not be buried in the dealer's scarto (discard).
 *
 * - `excuse` — the Excuse (Matto); never buriable.
 * - `bout` — a bout trump (Petit = trump 1, or the 21); never buriable.
 * - `trump` — an ordinary trump (atout); only buriable when unavoidable.
 * - `court` — a counting card (Valet/Cavalier/Dame/Roi, value ≥ 11); never buriable.
 */
export type ScartoUndiscardableReason = 'excuse' | 'bout' | 'trump' | 'court';

/** Minimum suit-card value that counts as a court (Valet 11 … Roi 14). Mirrors domain `ScartoCourtMin`. */
const SCARTO_COURT_MIN = 11;

/**
 * Classifies why a card cannot (normally) be buried in the Scarto discard, or
 * returns `null` for a freely buriable pip.
 *
 * Mirrors the backend rule (domain `validateScarto` / `scartoDiscardable`): the
 * Excuse and counting cards (Kings and courts, value ≥ 11) are never buriable,
 * the bouts (Petit = trump 1 and the 21) are never buriable, and ordinary trumps
 * may only be buried when the player holds fewer than three freely buriable pips.
 * Detection is color-based on the serialized tarot face — the Excuse is gold and
 * trumps (atouts) are purple, while a court is a non-tarot suit card of value ≥ 11.
 */
export function scartoUndiscardableReason(card: Card): ScartoUndiscardableReason | null {
  if (card.color === 'gold') return 'excuse';
  if (card.color === 'purple') {
    return card.value === 1 || card.value === 21 ? 'bout' : 'trump';
  }
  if (card.value >= SCARTO_COURT_MIN) return 'court';
  return null;
}
