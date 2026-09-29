import { formatGanjifaSuit, isGanjifaStrongSuit } from '../types/card';
import type { Card } from '../types/common';
import { GANJIFA_SUIT_GLYPHS } from '../types/games/ganjifa';

/** Returns the Ganjifa suit design represented by a card glyph, or null when unknown. */
export function ganjifaSuitOf(card: Card): number | null {
  const suit = GANJIFA_SUIT_GLYPHS.indexOf(card.glyph as (typeof GANJIFA_SUIT_GLYPHS)[number]);
  return suit > 0 ? suit : null;
}

/**
 * Screen-reader status for a Ganjifa hand card: its suit and whether higher or
 * lower numbers win in that suit. Returns undefined for a card without a known suit.
 */
export function ganjifaHandCardStatus(
  card: Card,
  t: (key: 'handCardStrong' | 'handCardWeak', params: { suit: string }) => string,
): string | undefined {
  const suit = ganjifaSuitOf(card);
  if (suit === null) return undefined;
  return t(isGanjifaStrongSuit(suit) ? 'handCardStrong' : 'handCardWeak', { suit: formatGanjifaSuit(suit) });
}
