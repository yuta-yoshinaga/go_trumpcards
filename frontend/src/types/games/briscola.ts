// Type declarations for briscola. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Briscola player data with points and trick count. */
export interface BriscolaPlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  points: number;
  trickCount: number;
}

/** A card played in a Briscola trick. */
export interface BriscolaTrickCard {
  playerIdx: number;
  card: Card;
}

/** Briscola game configuration. */
export interface BriscolaConfig {
  cpuDifficulty: number;
}

/** A suggested hint for Briscola. */
export interface BriscolaHint {
  cardIndex?: number;
  reason: string;
}

/** Full Briscola game state returned from the API. */
export interface BriscolaResponse extends BaseGameResponse {
  players: BriscolaPlayerData[];
  phase: number;
  trickNumber: number;
  currentPlayerIdx: number;
  currentTrick: BriscolaTrickCard[];
  trumpSuit: number;
  /** Face-up trump card (omitted once the stock is exhausted). */
  trumpCard?: Card;
  dealerIdx: number;
  leadPlayerIdx: number;
  /**
   * Cards remaining in the stock; this does NOT include the face-up trump
   * card (which is tracked separately via `trumpCard` until drawn last).
   */
  stockRemaining: number;
  gameEndFlag: boolean;
  /** -1 = tie or unfinished. */
  winnerIdx: number;
  config: BriscolaConfig;
  hint?: BriscolaHint;
}

// --- Schnapsen / Sixty-Six (シュナプセン / 66) ---
