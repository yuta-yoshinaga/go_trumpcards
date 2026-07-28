// Type declarations for pan. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** A meld (set or rope/run) laid on the table by a Panguingue player. */
export interface PanMeld {
  cards: Card[];
}

/** Panguingue player data with laid melds, chips, and scores. */
export interface PanPlayer {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  laidMelds: PanMeld[];
  meldedCount: number;
  chips: number;
  handPoints: number;
  roundScore: number;
  cumulativeScore: number;
}

/** Panguingue game configuration. */
export interface PanConfig {
  playerCount: number;
  cpuDifficulty: number;
  targetRounds: number;
}

/** Full Panguingue game state returned from the API. */
export interface PanResponse extends BaseGameResponse {
  players: PanPlayer[];
  phase: number;
  roundNumber: number;
  targetRounds: number;
  currentPlayerIdx: number;
  dealerIdx: number;
  discardTop: Card | null;
  drawPileCount: number;
  deckSize: number;
  winMeldCount: number;
  gameEndFlag: boolean;
  winnerIdx: number;
  panDeclarerIdx: number;
  config: PanConfig;
}

// --- Tonk (トンク) ---
