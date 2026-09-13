import type { Card, CardDesign } from '../types/card';

/** Suit order for the eight foundations (two decks), matching the domain definition. */
export const DIPLOMAT_FOUNDATION_SUITS: readonly CardDesign[] = [
  'SPADE',
  'CLOVER',
  'HEART',
  'DIAMOND',
  'SPADE',
  'CLOVER',
  'HEART',
  'DIAMOND',
] as const;

export const DIPLOMAT_FOUNDATION_TARGET = 13;

export interface DiplomatFoundationRequirement {
  suit: CardDesign;
  nextRank: number | null;
  canPlace: (card: Card | null | undefined) => boolean;
}

/**
 * Returns the required suit and next rank for a Diplomat foundation pile,
 * along with a predicate determining if a given card can be legally placed.
 *
 * Sync: Diplomat.canPlaceOnFoundation
 *
 * @param fIdx - The 0-based foundation index (0..7).
 * @param pile - The current cards in that foundation pile.
 * @returns An object with `suit`, `nextRank` (1..13 or `null` if completed), and `canPlace`.
 */
export function diplomatFoundationRequirement(fIdx: number, pile: readonly Card[]): DiplomatFoundationRequirement {
  const suit = DIPLOMAT_FOUNDATION_SUITS[fIdx] ?? 'SPADE';
  const pileLen = pile.length;
  if (pileLen >= DIPLOMAT_FOUNDATION_TARGET) {
    return {
      suit,
      nextRank: null,
      canPlace: () => false,
    };
  }
  // pileLen > 0 なら最後の要素は必ずあるので、既定値を置くとその枝は死ぬ。
  // noUncheckedIndexedAccess のために non-null 表明で受ける。
  const nextRank = pileLen === 0 ? 1 : pile[pileLen - 1]!.value + 1;
  return {
    suit,
    nextRank,
    canPlace: (card: Card | null | undefined): boolean => {
      if (!card) return false;
      return card.design === suit && card.value === nextRank;
    },
  };
}
