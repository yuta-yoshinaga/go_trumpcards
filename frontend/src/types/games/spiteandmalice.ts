// Type declarations for spiteandmalice. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** A single Spite & Malice player's view. Hand may contain `null` when
 * the opponent's cards are hidden from view. */
export interface SpiteAndMalicePlayerState {
  hand: (Card | null)[];
  goalTop?: Card;
  goalSize: number;
  sides: [Card[], Card[], Card[], Card[]];
  isCpu: boolean;
}

/** Hint information for the next recommended Spite & Malice move. */
export interface SpiteAndMaliceHint {
  source: 'goal' | 'hand' | 'side';
  index: number;
  foundationIdx: number;
  discard: boolean;
}

/** API response shape for a Spite & Malice game. */
export interface SpiteAndMaliceResponse extends BaseGameResponse {
  phase: number;
  current: number;
  players: [SpiteAndMalicePlayerState, SpiteAndMalicePlayerState];
  foundations: Card[][];
  foundationTops: number[];
  stockSize: number;
  completedSize: number;
  moveCount: number;
  winner: number;
  goalSize: number;
  cpuDifficulty: number;
  /** True when the human can auto-complete at least one foundation move on their turn. */
  canAutoComplete: boolean;
  hint?: SpiteAndMaliceHint;
}

/** Source or target zone for a Spite & Malice move. */
export interface SpiteAndMaliceMoveZone {
  zone: 'hand' | 'goal' | 'side' | 'foundation';
  idx?: number;
}

// --- Nertz / Pounce (ナーツ / パウンス) ---
