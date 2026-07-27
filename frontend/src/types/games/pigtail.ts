// Type declarations for pigtail. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Pig's Tail player output from the server. */
export interface PigsTailPlayer {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
}

/** Pig's Tail CPU action record. */
export interface PigsTailCpuAction {
  drawPlayerIdx: number;
  drawnCard: Card | null;
  penaltyFlag: boolean;
  penaltyCount: number;
  hesitationMs?: number;
}

/** Pig's Tail game state response. */
export interface PigsTailResponse extends BaseGameResponse {
  players: PigsTailPlayer[];
  circleCount: number;
  centerTop: Card | null;
  centerCount: number;
  currentTurn: number;
  gameEndFlag: boolean;
  loserIdx: number;
  lastDrawCard: Card | null;
  lastPenalty: boolean;
  cpuActions: PigsTailCpuAction[];
  humanAction: PigsTailCpuAction | null;
}

// --- Seven Card Stud ---
