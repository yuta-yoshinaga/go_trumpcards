// Type declarations for binokel. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';
import { BinokelPhase } from '../phases';

export { BinokelPhase };

/** Binokel game configuration. */
export interface BinokelConfig {
  cpuDifficulty: number;
  pointLimit: number;
}

/** Binokel meld data. */
export interface BinokelMeldData {
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
export interface BinokelMeldTableEntry {
  type: number;
  points: number;
}

/** Binokel trick card data. */
export interface BinokelTrickCard {
  playerIdx: number;
  card: Card;
}

/** Binokel player data (3 players, individual score). */
export interface BinokelPlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  score: number;
  trickCount: number;
  bid: number;
  hasPassed: boolean;
  meldScore: number;
  trickPoints: number;
}

/** Full Binokel game state returned from the API (matches BinokelWebController.go). */
export interface BinokelResponse extends BaseGameResponse {
  players: BinokelPlayerData[];
  phase: number;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  bidPlayerIdx: number;
  dealerIdx: number;
  trumpSuit: number;
  highestBid: number;
  highestBidder: number;
  currentTrick: BinokelTrickCard[];
  scores: [number, number, number] | number[];
  gameEndFlag: boolean;
  winnerPlayer: number;
  leadPlayerIdx: number;
  playerMelds: BinokelMeldData[][];
  dabb: Card[];
  dabbDiscarded: Card[];
  validPlayIndices?: number[];
  meldTable?: BinokelMeldTableEntry[];
  hint?: {
    cardIndex?: number;
    bidAmount?: number;
    pass?: boolean;
    suit?: number;
    reason: string;
  };
  config: BinokelConfig;
}
