// Type declarations for kalooki. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Kalooki meld: a set or run of cards laid face-up on the table. */
export interface KalookiMeld {
  cards: Card[];
}

/** Kalooki player state with face-up table melds, opening flag, and scores. */
export interface KalookiPlayer {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  melds: KalookiMeld[];
  /** Whether the player has made their opening meld(s) meeting the threshold. */
  hasOpened: boolean;
  roundScore: number;
  cumulativeScore: number;
}

/** Kalooki game configuration. */
export interface KalookiConfig {
  cpuDifficulty: number;
  playerCount: number;
  openingThreshold: number;
}

/** Kalooki API response. */
export interface KalookiResponse extends BaseGameResponse {
  players: KalookiPlayer[];
  /** 0 = draw, 1 = meld, 2 = round end, 3 = game end. */
  phase: number;
  /** Minimum points required for a player's opening meld. */
  openingThreshold: number;
  currentPlayerIdx: number;
  discardTop: Card | null;
  drawPileCount: number;
  gameEndFlag: boolean;
  winnerIdx: number;
  roundWinnerIdx: number;
  config: KalookiConfig;
}

// --- Oasis Poker (オアシスポーカー) ---
