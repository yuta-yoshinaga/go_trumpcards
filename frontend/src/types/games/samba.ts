// Type declarations for samba. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Samba game configuration. */
export interface SambaConfig {
  cpuDifficulty: number;
  pointLimit: number;
}

/**
 * A single meld on the table in Samba. `kind` distinguishes same-rank sets
 * (0) from suited sequences (1); `isCanasta`/`isSamba` flag the completed
 * seven-card set (canasta) and seven-card sequence (samba) respectively.
 */
export interface SambaMeldData {
  cards: Card[];
  kind: number; // 0 = set, 1 = sequence
  isNatural: boolean;
  isCanasta: boolean;
  isSamba: boolean;
  rank: number;
}

/** Samba player data with melds, red 3s, and partnership (team) affiliation. */
export interface SambaPlayerData {
  id: number;
  team: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  melds: SambaMeldData[];
  red3Count: number;
  red3s: Card[];
  roundScore: number;
  cumulativeScore: number;
  hasCanasta: boolean;
  hasSamba: boolean;
  hasInitMeld: boolean;
}

/** Full Samba game state returned from the API. */
export interface SambaResponse extends BaseGameResponse {
  players: SambaPlayerData[];
  teamScores: number[];
  phase: number;
  roundNumber: number;
  currentPlayerIdx: number;
  discardTop: Card | null;
  drawPileCount: number;
  discardPileCount: number;
  isFrozen: boolean;
  gameEndFlag: boolean;
  winnerIdx: number;
  config: SambaConfig;
}

// --- Hand and Foot (ハンド・アンド・フット) ---
