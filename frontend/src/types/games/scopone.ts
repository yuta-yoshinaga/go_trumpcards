// Type declarations for scopone. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Scopone player data (4 players in 2 teams). */
export interface ScoponePlayerData {
  id: number;
  isHuman: boolean;
  team: number;
  handCount: number;
  cards: Card[];
  capturedCount: number;
  scopaCount: number;
}

/** Scopone per-round score breakdown (per team, 2-element tuples). */
export interface ScoponeScoreDetail {
  cards: [number, number];
  diamonds: [number, number];
  sevens: [number, number];
  scopas: [number, number];
  gained: [number, number];
  settebello: number;
}

/** Scopone game rule configuration. */
export interface ScoponeConfig {
  cpuDifficulty: number;
  targetScore: number;
}

/** Full Scopone game state returned from the API. */
export interface ScoponeResponse extends BaseGameResponse {
  players: ScoponePlayerData[];
  tableCards: Card[];
  phase: string;
  roundNumber: number;
  currentTurn: number;
  dealerIdx: number;
  teamScores: number[];
  lastCaptureIdx: number;
  winnerTeam: number;
  gameEndFlag: boolean;
  isHumanTurn: boolean;
  /** Per human hand-card index, the list of valid table-index capture sets. */
  handCaptures: number[][][];
  lastRoundDetail?: ScoponeScoreDetail | null;
  config: ScoponeConfig;
}

// --- Escoba (エスコバ) ---
