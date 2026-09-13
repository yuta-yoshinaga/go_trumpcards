export const CARD_VALUE_MAX = 13;

/**
 * Returns the card value required for the n-th card (0-indexed) of a foundation pile.
 * Advances by 2 with wrap-around at 13 (CardValueMax).
 *
 * Sync: RoyalCotillion.royalCotillionNthValue
 *
 * @param base - The base starting rank (1 for Ace, 2 for Two).
 * @param n - The 0-based card index in the foundation pile (0..12).
 * @returns The required card rank (1..13).
 */
export function royalCotillionNthValue(base: number, n: number): number {
  return ((base - 1 + 2 * n) % CARD_VALUE_MAX) + 1;
}

/**
 * Returns the next rank required for a Royal Cotillion foundation pile,
 * or `null` if the foundation is complete (13 cards).
 *
 * @param pileLength - The number of cards currently in the foundation pile.
 * @param isOdd - True if this foundation starts with Ace (1), false if it starts with Two (2).
 * @returns The next required rank (1..13), or `null` if the pile is complete.
 */
export function royalCotillionNextRank(pileLength: number, isOdd: boolean): number | null {
  if (pileLength >= CARD_VALUE_MAX) return null;
  const base = isOdd ? 1 : 2;
  return royalCotillionNthValue(base, pileLength);
}
