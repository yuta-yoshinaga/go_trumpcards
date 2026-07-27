// Type declarations for pitch. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Pitch player data (4-player single-handed scoring). */
export interface PitchPlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  bid: number;
  roundScore: number;
  cumulativeScore: number;
  trickCount: number;
}

/** A card played in a Pitch trick. */
export interface PitchTrickCard {
  playerIdx: number;
  card: Card;
}

/** Pitch game configuration. */
export interface PitchConfig {
  cpuDifficulty: number;
  pointLimit: number;
}

/** A suggested hint for Pitch. */
export interface PitchHint {
  cardIndex?: number;
  bid?: number;
  reason: string;
}

/** Full Pitch game state returned from the API. */
export interface PitchResponse extends BaseGameResponse {
  players: PitchPlayerData[];
  phase: number;
  roundNumber: number;
  trickNumber: number;
  dealerIdx: number;
  currentPlayerIdx: number;
  bidPlayerIdx: number;
  currentBid: number;
  bidWinnerIdx: number;
  trumpSuit: number;
  currentTrick: PitchTrickCard[];
  /** Cards of the just-completed trick (empty on the round's first trick). */
  lastTrick: PitchTrickCard[];
  /** Winner index of the just-completed trick, or -1 when none. */
  lastTrickWinner: number;
  gameEndFlag: boolean;
  winnerIdx: number;
  leadPlayerIdx: number;
  validPlayIndices: number[];
  config: PitchConfig;
  hint?: PitchHint;
}

// --- Two Ten Jack (ツーテンジャック) ---
