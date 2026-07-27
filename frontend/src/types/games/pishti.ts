// Type declarations for pishti. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/**
 * Pişti game phase, mirroring the backend `PishtiPhase` string values
 * (internal/domain/Pishti.go). The phase is a string, not a numeric enum.
 */
export type PishtiPhase = 'play' | 'roundEnd' | 'gameEnd';

/** A single Pişti player as returned from the API. */
export interface PishtiPlayer {
  /** Seat index (0 = human). */
  id: number;
  isHuman: boolean;
  /** Number of cards currently in hand. */
  cardCount: number;
  /** The player's hand cards (populated only for the human). */
  cards: Card[];
  /** Total number of cards captured so far. */
  capturedCount: number;
  /** Accumulated Pişti bonus points. */
  pistiBonus: number;
  /** Final score (populated once the game ends). */
  finalScore: number;
}

/** Pişti configuration as returned from / sent to the API. */
export interface PishtiConfig {
  /** Number of players (2-4). */
  playerCnt: number;
  /** CPU difficulty (0=Easy, 1=Normal, 2=Hard). */
  cpuDifficulty: number;
}

/** Server response for the Pişti game (POST /pishti/exec). */
export interface PishtiResponse extends BaseGameResponse {
  players: PishtiPlayer[];
  /** Seat index whose turn it currently is. */
  currentTurn: number;
  /** All cards currently on the central pile, bottom to top. */
  pile: Card[];
  /** The top card of the pile, or null when the pile is empty. */
  pileTop: Card | null;
  /** Number of cards on the pile. */
  pileCount: number;
  /** Seat index of the most recent capturer, or -1. */
  lastCaptureIdx: number;
  gameEndFlag: boolean;
  /** Current game phase (a string, not a numeric enum). */
  phase: PishtiPhase | string;
  /** Cards remaining in the stock. */
  remainingDeck: number;
  /** Winning seat indices (may tie), empty until the game ends. */
  winners: number[];
  /** Final scores indexed by seat, empty until the game ends. */
  finalScores: number[];
  config: PishtiConfig;
}
