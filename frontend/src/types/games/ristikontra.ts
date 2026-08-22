// Type declarations for ristikontra. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/**
 * Pişti game phase, mirroring the backend `RistikontraPhase` string values
 * (internal/domain/Ristikontra.go). The phase is a string, not a numeric enum.
 */
export type RistikontraPhase = 'play' | 'roundEnd' | 'gameEnd';

/** A single Pişti player as returned from the API. */
export interface RistikontraPlayer {
  /** Seat index (0 = human). */
  id: number;
  isHuman: boolean;
  /** Number of cards currently in hand. */
  cardCount: number;
  /** The player's hand cards (populated only for the human). */
  cards: Card[];
  /** Total number of cards captured so far. */
  capturedCount: number;
  /** Final score (populated once the game ends). */
  finalScore: number;
}

/** Pişti configuration as returned from / sent to the API. */
export interface RistikontraConfig {
  /** Number of players (2-4). */
  playerCnt: number;
  /** CPU difficulty (0=Easy, 1=Normal, 2=Hard). */
  cpuDifficulty: number;
}

/** Server response for the Pişti game (POST /ristikontra/exec). */
export interface RistikontraResponse extends BaseGameResponse {
  players: RistikontraPlayer[];
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
  /**
   * Rank that the pending counter is on, or 0 when no capture is open.
   *
   * Playing this rank right now steals the bundle the last capturer just took.
   * The window closes as soon as anybody plays a different rank.
   */
  counterRank: number;
  gameEndFlag: boolean;
  /** Current game phase (a string, not a numeric enum). */
  phase: RistikontraPhase | string;
  /** Cards remaining in the stock. */
  remainingDeck: number;
  /** Winning seat indices (may tie), empty until the game ends. */
  winners: number[];
  /** Final scores indexed by seat, empty until the game ends. */
  finalScores: number[];
  config: RistikontraConfig;
}
