// Type declarations for pinochle. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Pinochle game configuration. */
export interface PinochleConfig {
  cpuDifficulty: number;
  pointLimit: number;
}

/** Pinochle meld data. */
export interface PinochleMeldData {
  type: number;
  points: number;
  cards: Card[];
}

/**
 * One row of the meld reference table: a meld type and what it scores.
 *
 * The server sends it with every state so the UI never carries a second copy
 * of the scoring values (#5519).
 */
export interface PinochleMeldTableEntry {
  type: number;
  points: number;
}

/** Pinochle trick card data. */
export interface PinochleTrickCard {
  playerIdx: number;
  card: Card;
}

/** Pinochle player data. */
export interface PinochlePlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  team: number;
  trickCount: number;
  bid: number;
  hasPassed: boolean;
  meldScore: number;
  trickPoints: number;
}

/** Full Pinochle game state returned from the API. */
export interface PinochleResponse extends BaseGameResponse {
  players: PinochlePlayerData[];
  phase: number;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  bidPlayerIdx: number;
  dealerIdx: number;
  trumpSuit: number;
  highestBid: number;
  highestBidder: number;
  currentTrick: PinochleTrickCard[];
  teamScores: [number, number];
  gameEndFlag: boolean;
  winnerTeam: number;
  leadPlayerIdx: number;
  playerMelds: PinochleMeldData[][];
  validPlayIndices?: number[];
  meldTable?: PinochleMeldTableEntry[];
  hint?: {
    cardIndex?: number;
    bidAmount?: number;
    pass?: boolean;
    suit?: number;
    reason: string;
  };
  config: PinochleConfig;
}

// --- Piquet (ピケ) ---
