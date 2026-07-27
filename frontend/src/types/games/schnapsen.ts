// Type declarations for schnapsen. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Schnapsen player data with points and trick count. */
export interface SchnapsenPlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  points: number;
  trickCount: number;
}

/** A card played in a Schnapsen trick. */
export interface SchnapsenTrickCard {
  playerIdx: number;
  card: Card;
}

/** Schnapsen game configuration. */
export interface SchnapsenConfig {
  cpuDifficulty: number;
}

/** A suggested hint for Schnapsen. */
export interface SchnapsenHint {
  cardIndex?: number;
  reason: string;
  isMarriage: boolean;
}

/** Full Schnapsen game state returned from the API. */
export interface SchnapsenResponse extends BaseGameResponse {
  players: SchnapsenPlayerData[];
  phase: number;
  trickNumber: number;
  currentPlayerIdx: number;
  currentTrick: SchnapsenTrickCard[];
  trumpSuit: number;
  /** Face-up trump upcard (omitted once the stock is exhausted). */
  trumpCard?: Card;
  dealerIdx: number;
  leadPlayerIdx: number;
  /** Cards remaining in the stock (excludes the face-up trump upcard). */
  stockRemaining: number;
  /** True once the stock is exhausted and must-follow rules apply (phase 2). */
  isEndgame: boolean;
  /** Indices in the human hand that are legal to play right now. */
  validPlays: number[];
  /** Indices in the human hand that can start a marriage declaration. */
  marriagePlays: number[];
  gameEndFlag: boolean;
  /** -1 = tie or unfinished. */
  winnerIdx: number;
  config: SchnapsenConfig;
  hint?: SchnapsenHint;
}

// --- Truco (トゥルコ) ---
