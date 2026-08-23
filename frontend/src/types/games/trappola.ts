// Type declarations for trappola. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** A Trappola player's public/own state. */
export interface TrappolaPlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  trickCount: number;
  teamId: number;
}

/** A card played in a Trappola trick. */
export interface TrappolaTrickCard {
  playerIdx: number;
  card: Card;
}

/** Trappola game configuration. */
export interface TrappolaConfig {
  cpuDifficulty: number;
  targetPoints: number;
}

/** A suggested hint for Trappola. */
export interface TrappolaHint {
  cardIndices: number[];
  reason: string;
}

/** Full Trappola game state returned from the API. */
export interface TrappolaResponse extends BaseGameResponse {
  players: TrappolaPlayerData[];
  phase: number;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  currentTrick: TrappolaTrickCard[];
  lastTrick: TrappolaTrickCard[];
  lastTrickWinner: number;
  leadPlayerIdx: number;
  teamScores: number[];
  teamRoundThirds: number[];
  playableIndices: number[];
  gameEndFlag: boolean;
  winnerTeam: number;
  config: TrappolaConfig;
  hint?: TrappolaHint;
}

// --- Sheepshead ---
