// Type declarations for doubt. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Bracket data for doubt profile export. */
export interface DoubtProfileBracketData {
  bluffs: number;
  total: number;
}

/** Exported doubt human profile data. */
export interface DoubtHumanProfileData {
  bluffsByBracket: [DoubtProfileBracketData, DoubtProfileBracketData, DoubtProfileBracketData];
  doubtCorrect: number;
  doubtTotal: number;
  gamesPlayed: number;
  hesitationCount: number;
  hesitationMean: number;
  hesitationM2: number;
}

/** Doubt player data with card count and finish status. */
export interface DoubtPlayerData {
  id: number;
  isHuman: boolean;
  isFinished: boolean;
  cardCount: number;
  cards: Card[];
}

/** CPU play action in Doubt with bluff information. */
export interface DoubtCpuAction {
  playerIdx: number;
  claimedValue: number;
  cardCount: number;
  isBluff: boolean;
  hasTell?: boolean;
  hesitationMs?: number;
}

/** Result of a doubt challenge in Doubt. */
export interface DoubtDoubtResult {
  doubterIdx: number;
  cardPlayerIdx: number;
  wasLying: boolean;
  loserIdx: number;
  cardCount: number;
  discardedCount: number;
  revealedCards: Card[];
}

/** Doubt game configuration options. */
export interface DoubtConfig {
  doubtWindowSec: number;
  cpuMemoryLevel: number; // 0=Easy, 1=Normal, 2=Hard
  penaltyDrawLimit: number; // 0=unlimited, >0=max cards loser picks up
  cpuHesitationEnabled: boolean;
  cpuMetaAI: boolean;
}

/** Full Doubt game state returned from the API. */
export interface DoubtResponse extends BaseGameResponse {
  players: DoubtPlayerData[];
  currentTurn: number;
  phase: 0 | 1 | 2; // 0=Play, 1=Doubt, 2=End
  tableCardCount: number;
  lastAction: DoubtCpuAction | null;
  cpuDoubters: number[];
  cpuActions: DoubtCpuAction[];
  humanAction: DoubtCpuAction | null;
  lastDoubtResult: DoubtDoubtResult | null;
  gameEndFlag: boolean;
  winnerIdx: number;
  doubtWindowSec: number;
  penaltyDrawLimit: number;
  metaAI?: DoubtMetaAI;
  profile?: DoubtHumanProfileData;
}

/** Meta-AI statistics for Doubt CPU adaptation. */
export interface DoubtMetaAI {
  enabled: boolean;
  gamesPlayed: number;
  bluffRate: number;
  doubtAccuracy: number;
  hesitationMean: number;
}
