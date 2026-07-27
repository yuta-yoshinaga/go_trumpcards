// Type declarations for oldmaid. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Exported old maid human profile data. */
export interface OldMaidHumanProfileData {
  positionBuckets: [number, number, number];
  totalPicks: number;
  shuffleCount: number;
  drawCount: number;
  gamesPlayed: number;
}

/** Old Maid player data with hand and finish status. */
export interface OldMaidPlayerData {
  id: number;
  isHuman: boolean;
  isFinished: boolean;
  cardCount: number;
  cards: Card[];
}

/** Meta-AI statistics for Old Maid CPU adaptation. */
export interface OldMaidMetaAI {
  enabled: boolean;
  gamesPlayed: number;
  edgePickRate: number;
}

/** Full Old Maid game state returned from the API. */
export interface OldMaidResponse extends BaseGameResponse {
  players: OldMaidPlayerData[];
  currentTurn: number;
  nextDrawTargetIdx: number;
  gameEndFlag: boolean;
  hasDrawn: boolean;
  lastDrawPlayerIdx: number;
  lastDrawFromIdx: number;
  lastDrawCard: Card | null;
  lastDiscardedPairs: number;
  lastDiscardedCards?: Card[];
  cpuActions: CpuAction[];
  humanAction?: CpuAction | null;
  drawHistory: DrawHistoryEntry[];
  cpuHighlightedCardIdx: number;
  removedCard: Card | null;
  mode: number;
  metaAI?: OldMaidMetaAI;
  profile?: OldMaidHumanProfileData;
}

/** CPU draw/discard action in Old Maid. */
export interface CpuAction {
  drawPlayerIdx: number;
  drawFromIdx: number;
  drawnCard: Card | null;
  discardedPairs: number;
  discardedCards?: Card[];
  hesitationMs?: number;
}

/** History entry for a card draw in Old Maid. */
export interface DrawHistoryEntry {
  drawPlayerIdx: number;
  drawFromIdx: number;
  discardedPairs: number;
  drawerFinished: boolean;
  targetFinished: boolean;
}
