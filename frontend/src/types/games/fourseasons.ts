// Type declarations for fourseasons. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** A suggested move hint in Four Seasons. */
export interface FourSeasonsHint {
  fromZone: string;
  /** Cross pile index, or -1 when the source is the waste. */
  fromIdx: number;
  toZone: string;
  toIdx: number;
}

/** Full Four Seasons game state returned from the API. */
export interface FourSeasonsResponse extends BaseGameResponse {
  tableau: Card[][];
  foundation: Card[][];
  stockCount: number;
  waste: Card[];
  /**
   * The rank the foundations start from, fixed by the first card of this deal
   * (1..13). Every placement rule reads it, so the page must not assume Ace.
   */
  baseRank: number;
  phase: number;
  moveCount: number;
  canUndo: boolean;
  hint?: FourSeasonsHint;
}

/** Source or target zone for a Four Seasons card move. */
export interface FourSeasonsMoveZone {
  zone: 'waste' | 'tableau' | 'foundation';
  idx?: number;
}
