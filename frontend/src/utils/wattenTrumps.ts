import type { Card, CardDesign } from '../types/card';

/**
 * The trump category a Watten card falls into once a Schlag rank and critical
 * (trump) suit are declared. Ordered from strongest to weakest tier:
 *
 * - `max`      — ♥K, the permanent highest trump ("Max"/Guater).
 * - `belli`    — ♦K, the permanent second trump.
 * - `spitz`    — ♦7, the permanent third trump.
 * - `schlag`   — any card of the declared Schlag rank (except the three above).
 * - `critical` — any remaining card of the declared critical suit.
 *
 * Mirrors the ranking in `internal/domain/Watten.go` (`cardRank`/`isTrump`).
 */
export type WattenTrumpCategory = 'max' | 'belli' | 'spitz' | 'schlag' | 'critical';

/** A card in the human's hand that is a trump, with its category and strength. */
export interface WattenTrumpCard {
  /** The card itself. */
  card: Card;
  /** Index of the card within the source hand array. */
  index: number;
  /** Which trump tier the card belongs to. */
  category: WattenTrumpCategory;
  /** Trick-comparison strength (higher = stronger), mirroring the domain. */
  rank: number;
}

/** Maps a numeric critical-suit code (1=♠ 2=♣ 3=♥ 4=♦) to a card design. */
const SUIT_CODE_TO_DESIGN: Readonly<Record<number, CardDesign>> = {
  1: 'SPADE',
  2: 'CLOVER',
  3: 'HEART',
  4: 'DIAMOND',
};

/** Fixed suit order for tie-breaking Schlag cards: ♥>♦>♠>♣ (mirrors domain). */
function schlagSuitOrder(design: CardDesign): number {
  switch (design) {
    case 'HEART':
      return 3;
    case 'DIAMOND':
      return 2;
    case 'SPADE':
      return 1;
    default:
      return 0;
  }
}

/** Value strength A>K>Q>J>10>9>8>7 for critical-suit / plain cards (mirrors domain). */
function wattenValueRank(value: number): number {
  switch (value) {
    case 1: // A
      return 8;
    case 13: // K
      return 7;
    case 12: // Q
      return 6;
    case 11: // J
      return 5;
    case 10:
      return 4;
    case 9:
      return 3;
    case 8:
      return 2;
    case 7:
      return 1;
    default:
      return 0;
  }
}

/**
 * Classifies a single card given the declared Schlag rank and critical suit,
 * returning its trump category and trick-comparison strength, or `null` when
 * the card is a plain (non-trump) card.
 *
 * A Schlag rank of 0 (or a critical suit outside 1..4) is treated as "not yet
 * declared" for that dimension, so the classification updates live as the
 * dealer picks each choice. The three permanent trumps (Max ♥K, Belli ♦K,
 * Spitz ♦7) are trumps regardless of the declaration.
 *
 * @param card - The card to classify.
 * @param schlagRank - Declared Schlag rank (1=A, 7..13), or 0 when unset.
 * @param criticalSuit - Declared critical suit code (1=♠ 2=♣ 3=♥ 4=♦), or 0/-1 when unset.
 * @returns The card's category and rank, or `null` when it is not a trump.
 */
export function wattenTrumpInfo(
  card: Card,
  schlagRank: number,
  criticalSuit: number,
): { category: WattenTrumpCategory; rank: number } | null {
  if (card.design === 'HEART' && card.value === 13) return { category: 'max', rank: 1000 };
  if (card.design === 'DIAMOND' && card.value === 13) return { category: 'belli', rank: 999 };
  if (card.design === 'DIAMOND' && card.value === 7) return { category: 'spitz', rank: 998 };
  if (schlagRank > 0 && card.value === schlagRank) {
    return { category: 'schlag', rank: 900 + schlagSuitOrder(card.design) };
  }
  if (SUIT_CODE_TO_DESIGN[criticalSuit] === card.design) {
    return { category: 'critical', rank: 800 + wattenValueRank(card.value) };
  }
  return null;
}

/**
 * Returns the trump cards within a hand for the given Schlag rank and critical
 * suit, sorted strongest-first, so a Watten dealer can preview which of their
 * cards become the top trumps ("critical" cards) before declaring.
 *
 * @param cards - The hand to scan.
 * @param schlagRank - Declared Schlag rank (1=A, 7..13), or 0 when unset.
 * @param criticalSuit - Declared critical suit code (1=♠ 2=♣ 3=♥ 4=♦), or 0/-1 when unset.
 * @returns Trump cards with their index, category, and strength, strongest first.
 */
export function wattenTrumpCards(cards: Card[], schlagRank: number, criticalSuit: number): WattenTrumpCard[] {
  const trumps: WattenTrumpCard[] = [];
  cards.forEach((card, index) => {
    const info = wattenTrumpInfo(card, schlagRank, criticalSuit);
    if (info) trumps.push({ card, index, category: info.category, rank: info.rank });
  });
  return trumps.sort((a, b) => b.rank - a.rank);
}
