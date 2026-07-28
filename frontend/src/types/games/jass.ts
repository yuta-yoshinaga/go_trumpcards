// Type declarations for jass. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Jass player data with team, trick count, and hand. */
export interface JassPlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  team: number;
  trickCount: number;
}

/** A card played in a Jass trick. */
export interface JassTrickCard {
  playerIdx: number;
  card: Card;
}

/** Jass game configuration. */
export interface JassConfig {
  cpuDifficulty: number;
  targetScore: number;
  lastTrickBonus: number;
  enableWeis: boolean;
}

/** A suggested hint for Jass. */
export interface JassHint {
  cardIndex?: number;
  schieben?: boolean;
  suit?: number;
  reason: string;
}

/** Full Jass (Schieber) game state returned from the API. */
export interface JassResponse extends BaseGameResponse {
  players: JassPlayerData[];
  phase: number;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  bidPlayerIdx: number;
  dealerIdx: number;
  forehandIdx: number;
  trumpSuit: number;
  schieben: boolean;
  makerTeam: number;
  makerPlayerIdx: number;
  currentTrick: JassTrickCard[];
  lastTrick: JassTrickCard[];
  lastTrickWinner: number;
  teamScores: number[];
  roundPoints: number[];
  roundWeisPoints: number[];
  roundStockPoints: number[];
  gameEndFlag: boolean;
  winnerTeam: number;
  leadPlayerIdx: number;
  config: JassConfig;
  hint?: JassHint;
}

// --- Watten (ヴァッテン) ---
