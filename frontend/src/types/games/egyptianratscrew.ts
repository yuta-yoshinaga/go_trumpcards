// Type declarations for egyptianratscrew. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Player snapshot for Egyptian Ratscrew. */
export interface EgyptianRatscrewPlayerData {
  name: string;
  isHuman: boolean;
  stockSize: number;
}

/** Full Egyptian Ratscrew game state returned from the API. */
export interface EgyptianRatscrewResponse extends BaseGameResponse {
  phase: number;
  gameEndFlag: boolean;
  winnerIdx: number;
  currentTurnIdx: number;
  isHumanTurn: boolean;
  isTopFaceCard: boolean;
  isSlappable: boolean;
  centerPileSize: number;
  topCard?: Card | null;
  players: EgyptianRatscrewPlayerData[];
  cpuDifficulty: number;
  chanceRemaining: number;
  chanceFromIdx: number;
  pendingKind: number;
  pendingDeadlineMs: number;
  lastEventKind: number;
  lastEventPlayerIdx: number;
  lastSlapReason: number;
}

// --- Contract Rummy (コントラクトラミー) ---
