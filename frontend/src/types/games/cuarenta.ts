// Type declarations for cuarenta. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** A single Cuarenta player as returned from the API. */
export interface CuarentaPlayer {
  /** Seat index (0 = human; seats {0,2}=Team A, {1,3}=Team B). */
  id: number;
  /** Team index (0 = Team A, 1 = Team B). */
  team: number;
  isHuman: boolean;
  /** Number of cards currently in hand. */
  cardCount: number;
  /** The player's hand cards (populated only for the human). */
  cards: Card[];
  /** Total number of cards captured by this player so far this round. */
  capturedCount: number;
}

/** A single Cuarenta play action (human or CPU), describing what was captured. */
export interface CuarentaAction {
  /** Seat index of the acting player. */
  playerIdx: number;
  /** The card that was played, or null. */
  playedCard: Card | null;
  /** Cards captured by this play (empty when the card was laid on the table). */
  capturedCards: Card[];
  /** True when this play scored a caída (+2). */
  isCaida: boolean;
  /** True when this play cleared the table (limpia, +1). */
  isLimpia: boolean;
  /** Extra ronda points scored by this play (0 when none). */
  rondaBonus: number;
}

/** Round-end scoring breakdown keyed by team index. */
export interface CuarentaScoreDetail {
  /** Cards captured this round, keyed by team index. */
  capturedCount: Record<string, number>;
  /** Caída points, keyed by team index. */
  caida: Record<string, number>;
  /** Ronda points, keyed by team index. */
  ronda: Record<string, number>;
  /** Limpia points, keyed by team index. */
  limpia: Record<string, number>;
  /** Team index awarded the most-cards bonus, or -1 when none. */
  mostCards: number;
  /** Total points gained this round, keyed by team index. */
  gained: Record<string, number>;
}

/** Cuarenta configuration as returned from / sent to the API. */
export interface CuarentaConfig {
  /** Target score to win the game (default 40). */
  targetScore: number;
  /** CPU difficulty (0=Easy, 1=Normal, 2=Hard). */
  cpuDifficulty: number;
}

/** Server response for the Cuarenta game (POST /cuarenta/exec). */
export interface CuarentaResponse extends BaseGameResponse {
  players: CuarentaPlayer[];
  /** Seat index whose turn it currently is. */
  currentTurn: number;
  /** All cards currently on the central table. */
  tableCards: Card[];
  /** Seat index of the most recent capturer, or -1. */
  lastCaptureIdx: number;
  gameEndFlag: boolean;
  /** Current game phase (0=Play, 1=RoundEnd, 2=GameEnd). */
  phase: number;
  /** Cumulative score per team, indexed by team. */
  teamScores: number[];
  /** Cards remaining in the stock. */
  remainingDeck: number;
  /** Winning team indices (may tie), empty until the game ends. */
  roundWinners: number[];
  /** CPU actions that occurred since the last human play. */
  cpuActions: CuarentaAction[];
  /** The human's most recent action, or null. */
  humanAction: CuarentaAction | null;
  /** The most recent round-end scoring breakdown, or null. */
  lastRoundDetail: CuarentaScoreDetail | null;
  config: CuarentaConfig;
}
