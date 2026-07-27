// Type declarations for slapjack. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Player snapshot for Slapjack. */
export interface SlapjackPlayerData {
  name: string;
  isHuman: boolean;
  stockSize: number;
}

/** Full Slapjack game state returned from the API. */
export interface SlapjackResponse extends BaseGameResponse {
  phase: number;
  gameEndFlag: boolean;
  winnerIdx: number;
  currentTurnIdx: number;
  isHumanTurn: boolean;
  isTopJack: boolean;
  centerPileSize: number;
  topCard?: Card | null;
  players: SlapjackPlayerData[];
  cpuDifficulty: number;
  pendingKind: number;
  pendingDeadlineMs: number;
  lastEventKind: number;
  lastEventPlayerIdx: number;
}
