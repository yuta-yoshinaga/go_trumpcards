// Type declarations for war. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** War player data with face-down pile and discard pile sizes. */
export interface WarPlayerData {
  id: number;
  isHuman: boolean;
  drawPileSize: number;
  discardPileSize: number;
  totalCards: number;
}

/** War game configuration. */
export interface WarConfig {
  maxRounds: number;
}

/** Full War game state returned from the API. */
export interface WarResponse extends BaseGameResponse {
  players: WarPlayerData[];
  phase: number;
  gameEndFlag: boolean;
  winnerIdx: number;
  playerRevealed: Card | null;
  cpuRevealed: Card | null;
  warPotSize: number;
  lastWinnerIdx: number;
  lastBurialCount: number;
  roundsPlayed: number;
  config: WarConfig;
}
