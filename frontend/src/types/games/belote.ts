// Type declarations for belote. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Belote player data with team, trick count, and hand. */
export interface BelotePlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  team: number;
  trickCount: number;
}

/** A card played in a Belote trick. */
export interface BeloteTrickCard {
  playerIdx: number;
  card: Card;
}

/** Belote game configuration. */
export interface BeloteConfig {
  cpuDifficulty: number;
  targetScore: number;
  dixDeDer: number;
  enableBeloteRebelote: boolean;
}

/** A suggested hint for Belote. */
export interface BeloteHint {
  cardIndex?: number;
  orderUp?: boolean;
  suit?: number;
  reason: string;
}

/** Full Belote game state returned from the API. */
export interface BeloteResponse extends BaseGameResponse {
  players: BelotePlayerData[];
  phase: number;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  bidPlayerIdx: number;
  dealerIdx: number;
  trumpSuit: number;
  faceUpCard: Card | null;
  makerTeam: number;
  makerPlayerIdx: number;
  currentTrick: BeloteTrickCard[];
  teamScores: number[];
  roundPoints: number[];
  roundBeloteBonus: number[];
  gameEndFlag: boolean;
  winnerTeam: number;
  leadPlayerIdx: number;
  config: BeloteConfig;
  hint?: BeloteHint;
}

// --- Jass (Schieber) (ヤス) ---
