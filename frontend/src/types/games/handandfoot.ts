// Type declarations for handandfoot. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Hand and Foot game configuration. */
export interface HandAndFootConfig {
  cpuDifficulty: number;
  pointLimit: number;
}

/** A single team meld on the table in Hand and Foot. */
export interface HandAndFootMeldData {
  cards: Card[];
  isNatural: boolean;
  isCanasta: boolean;
  rank: number;
}

/** Per-team meld and red-3 data in Hand and Foot. */
export interface HandAndFootTeamData {
  team: number;
  melds: HandAndFootMeldData[];
  red3Count: number;
  red3s: Card[];
}

/** Hand and Foot player data. Melds and red 3s are held per team, not per player. */
export interface HandAndFootPlayerData {
  id: number;
  team: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  footCount: number;
  inFoot: boolean;
  roundScore: number;
  cumulativeScore: number;
}

/** Full Hand and Foot game state returned from the API. */
export interface HandAndFootResponse extends BaseGameResponse {
  players: HandAndFootPlayerData[];
  teams: HandAndFootTeamData[];
  phase: number;
  roundNumber: number;
  currentPlayerIdx: number;
  discardTop: Card | null;
  drawPileCount: number;
  discardPileCount: number;
  isFrozen: boolean;
  gameEndFlag: boolean;
  winnerTeam: number;
  config: HandAndFootConfig;
}

// --- Burraco (ブラーコ) ---
