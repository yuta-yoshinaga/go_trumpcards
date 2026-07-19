import type { Card } from '../types/card';

/**
 * The three French Tarot bouts (oudlers): the 21 (highest trump), the Petit (trump 1),
 * and the Excuse (the Fool). The number of bouts a side captures sets the card-point
 * target required for a made contract, so knowing which bouts you hold is central to play.
 */
export type FrenchTarotBout = 'twentyOne' | 'petit' | 'excuse';

/**
 * True when the card is a trump (atout). Trumps are drawn in `purple` by the French Tarot
 * face descriptor (see `FrenchTarotWebPresenter.frenchTarotFace`); their `design` field is
 * the generic "JOKER" sentinel, so `color` is the reliable discriminator.
 */
function isTrump(card: Card): boolean {
  return card.color === 'purple';
}

/** True when the card is the Excuse (the Fool), drawn in `gold`. */
function isExcuse(card: Card): boolean {
  return card.color === 'gold';
}

/** The three bouts in canonical display order. */
const BOUT_ORDER: readonly FrenchTarotBout[] = ['twentyOne', 'petit', 'excuse'];

/** True when the given card is the named bout. */
function cardIsBout(card: Card, bout: FrenchTarotBout): boolean {
  switch (bout) {
    case 'twentyOne':
      return isTrump(card) && card.value === 21;
    case 'petit':
      return isTrump(card) && card.value === 1;
    case 'excuse':
      return isExcuse(card);
  }
}

/**
 * Returns which bouts are present in the given hand, in canonical display order
 * (21, Petit, Excuse). Note the Petit is the trump valued 1 — a suit Ace (also valued 1
 * but not purple) is correctly excluded.
 *
 * @param cards - The hand to inspect.
 * @returns The bouts held, in canonical order (possibly empty).
 */
export function heldBouts(cards: readonly Card[]): FrenchTarotBout[] {
  return BOUT_ORDER.filter((bout) => cards.some((card) => cardIsBout(card, bout)));
}

/**
 * The card-point target required for a made contract given the number of bouts the declarer
 * captures (0→56, 1→51, 2→41, 3→36). Mirrors the domain `frenchTarotTarget`.
 *
 * @param bouts - The number of bouts captured (0-3).
 * @returns The required card-point target (half-points).
 */
export function frenchTarotTarget(bouts: number): number {
  if (bouts <= 0) return 56;
  if (bouts === 1) return 51;
  if (bouts === 2) return 41;
  return 36;
}
