// Type declarations for escoba. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Escoba player data (4 players, free-for-all / no teams). */
export interface EscobaPlayerData {
  id: number;
  isHuman: boolean;
  handCount: number;
  cards: Card[];
  capturedCount: number;
  /** The captured pile's actual cards. Populated only for the human player; CPUs stay count-only (empty array). */
  capturedCards: Card[];
  escobaCount: number;
  score: number;
}

/**
 * Escoba per-round score breakdown (per-player arrays, one entry per player).
 * `aceEspada` / `seteEspada` are the player indices who took the Ace♠ and 7♠.
 */
export interface EscobaScoreDetail {
  cards: number[];
  espadas: number[];
  sevens: number[];
  oros: number[];
  escobas: number[];
  gained: number[];
  aceEspada: number;
  seteEspada: number;
}

/** Escoba game rule configuration. */
export interface EscobaConfig {
  cpuDifficulty: number;
  targetScore: number;
}

/** Full Escoba game state returned from the API. */
export interface EscobaResponse extends BaseGameResponse {
  players: EscobaPlayerData[];
  tableCards: Card[];
  phase: string;
  roundNumber: number;
  currentTurn: number;
  dealerIdx: number;
  stockRemaining: number;
  lastCaptureIdx: number;
  winnerIdx: number;
  gameEndFlag: boolean;
  isHumanTurn: boolean;
  /** Per human hand-card index, the list of valid table-index capture sets summing to 15. */
  handCaptures: number[][][];
  lastRoundDetail?: EscobaScoreDetail | null;
  config: EscobaConfig;
}

// --- Barbu (バルブ) ---
