// Type declarations for gaigel. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Gaigel player data with team, trick count, and hand. */
export interface GaigelPlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  team: number;
  trickCount: number;
}

/** A card played in a Gaigel trick. */
export interface GaigelTrickCard {
  playerIdx: number;
  card: Card;
}

/** Gaigel game configuration. */
export interface GaigelConfig {
  cpuDifficulty: number;
  targetScore: number;
}

/** A suggested hint for Gaigel. */
export interface GaigelHint {
  cardIndex?: number;
  reason: string;
  isMarriage: boolean;
}

/** Full Gaigel game state returned from the API. */
export interface GaigelResponse extends BaseGameResponse {
  players: GaigelPlayerData[];
  phase: number;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  dealerIdx: number;
  trumpSuit: number;
  trumpCard?: Card;
  stockRemaining: number;
  /**
   * Whether the stock is exhausted and must-follow has taken over.
   *
   * **This is a rule change, not a counter.** From this point `validatePlay`
   * enforces following the led suit and trumping; the sister games (Bezique,
   * Schnapsen) both surface it, and Gaigel was the only one that did not
   * (#6482).
   */
  isEndgame: boolean;
  currentTrick: GaigelTrickCard[];
  teamScores: number[];
  roundPoints: number[];
  roundMarriage: number[];
  marriageIndices: number[];
  gameEndFlag: boolean;
  winnerTeam: number;
  leadPlayerIdx: number;
  config: GaigelConfig;
  hint?: GaigelHint;
}

// --- Contract Bridge (コントラクトブリッジ) ---
