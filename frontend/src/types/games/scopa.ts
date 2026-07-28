// Type declarations for scopa. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Scopa player data. */
export interface ScopaPlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  capturedCount: number;
  scopaCount: number;
  totalScore: number;
}

/** A play/lay action in Scopa. */
export interface ScopaAction {
  playerIdx: number;
  playedCard: Card | null;
  capturedCards: Card[];
  isScopa: boolean;
}

/** Scopa score detail (per round). */
export interface ScopaScoreDetail {
  cards: Record<number, number>;
  diamonds: Record<number, number>;
  sevens: Record<number, number>;
  hasSetteBello: number;
  scopas: Record<number, number>;
  gained: Record<number, number>;
}

/** Scopa game rule configuration. */
export interface ScopaConfig {
  targetScore: number;
  cpuDifficulty: number;
}

/** Full Scopa game state returned from the API. */
export interface ScopaResponse extends BaseGameResponse {
  players: ScopaPlayerData[];
  currentTurn: number;
  tableCards: Card[];
  lastCaptureIdx: number;
  gameEndFlag: boolean;
  phase: string;
  config: ScopaConfig;
  cpuActions: ScopaAction[];
  humanAction: ScopaAction | null;
  remainingDeck: number;
  packsDealt: number;
  roundWinners: number[];
  lastRoundDetail: ScopaScoreDetail | null;
}

// --- Scopone (スコポーネ) ---
