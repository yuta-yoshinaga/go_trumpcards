import type { Card } from '../types/card';

/**
 * The named ranks of Karnöffel's chosen suit, mirroring the constants in
 * `internal/domain/Karnoffel.go`. Only cards of the chosen suit hold a title;
 * the same rank in any other suit is an ordinary card.
 */
export const KARNOFFEL_RANK_KEYS: Readonly<Record<number, string>> = {
  11: 'karnoffel',
  7: 'devil',
  6: 'pope',
  2: 'kaiser',
  3: 'oberstecher',
  4: 'unterstecher',
  5: 'farbenstecher',
};

/**
 * i18n key suffix of the title `card` holds, or null when it holds none.
 *
 * Karnöffel's whole difficulty is that its order is irregular and depends on a
 * suit chosen each hand, so which card in front of you is the Pope is not
 * something the static ladder text can answer (#4773).
 * @param card - The card to name.
 * @param chosenSuit - The suit chosen for this hand.
 * @returns The rank key, or null for a plain card.
 */
export function karnoffelRankKey(card: Card, chosenSuit: number): string | null {
  if (card.value === undefined) return null;
  if (suitIndex(card.design) !== chosenSuit) return null;
  return KARNOFFEL_RANK_KEYS[card.value] ?? null;
}

/** Card designs in the numeric order the backend uses (1-based, matching Card.GetDesign). */
const DESIGN_ORDER: Readonly<Record<string, number>> = { SPADE: 1, CLOVER: 2, HEART: 3, DIAMOND: 4 };

/**
 * Numeric suit index for a card design, or -1 for a design with no suit (joker).
 * @param design - The card design.
 * @returns The suit index.
 */
export function suitIndex(design: string): number {
  return DESIGN_ORDER[design] ?? -1;
}
