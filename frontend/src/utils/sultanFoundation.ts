import type { Card } from '../types/card';
import { suitSymbol } from './cardAlt';

/**
 * Number of cards on a completed Sultan foundation (King + A..Q = 13).
 * Mirrors the backend `domain.SultanFoundationFull`.
 */
export const SULTAN_FOUNDATION_FULL = 13;

/** Display metadata for a single Sultan foundation pile. */
export interface SultanFoundationInfo {
  /** Suit symbol (♠♥♦♣) of the King base, or an empty string when the pile has no base. */
  suit: string;
  /** Suit design of the King base, or null when the pile is empty. */
  design: Card['design'] | null;
  /** Number of cards on the pile (>=1 once the King base is placed). */
  count: number;
  /** True once the pile holds all 13 cards (K + A..Q). */
  complete: boolean;
}

/**
 * Derive display metadata for a Sultan foundation pile.
 *
 * Every Sultan foundation is seeded with a King (two decks → two piles per
 * suit) and builds K→A→2…→Q in that suit, so the pile's target suit is read
 * from its base card (`pile[0]`). An empty pile (no King base) reports no suit.
 */
export function sultanFoundationInfo(pile: Card[]): SultanFoundationInfo {
  const base = pile.length > 0 ? pile[0] : null;
  return {
    suit: base ? suitSymbol(base.design) : '',
    design: base ? base.design : null,
    count: pile.length,
    complete: pile.length >= SULTAN_FOUNDATION_FULL,
  };
}
