// Type declarations for bridge. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Bridge player data with team, trick count, and hand. */
export interface BridgePlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  team: number;
  trickCount: number;
}

/** A card played in a Bridge trick. */
export interface BridgeTrickCard {
  playerIdx: number;
  card: Card;
}

/** A bid entry in the Bridge bid history. */
export interface BridgeBidEntry {
  playerIdx: number;
  bidType: number;
  level: number;
  suit: number;
}

/** Bridge game configuration. */
export interface BridgeConfig {
  cpuDifficulty: number;
}

/** A suggested hint for Bridge. */
export interface BridgeHint {
  cardIndex?: number;
  bidType?: number;
  bidLevel?: number;
  bidSuit?: number;
  reason: string;
}

/** Full Bridge game state returned from the API. */
export interface BridgeResponse extends BaseGameResponse {
  players: BridgePlayerData[];
  phase: number;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  bidPlayerIdx: number;
  dealerIdx: number;
  trumpSuit: number;
  contractLevel: number;
  contractSuit: number;
  doubled: number;
  declarerIdx: number;
  dummyIdx: number;
  bidHistory: BridgeBidEntry[];
  vulnerability: boolean[];
  currentTrick: BridgeTrickCard[];
  teamScores: number[];
  gamesWon: number[];
  belowLine: number[];
  gameEndFlag: boolean;
  winnerTeam: number;
  leadPlayerIdx: number;
  openingLeadDone: boolean;
  dummyHand: Card[] | null;
  config: BridgeConfig;
  hint?: BridgeHint;
}

// --- Pyramid Solitaire (ピラミッド) ---
