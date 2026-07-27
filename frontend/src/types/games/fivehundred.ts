// Type declarations for fivehundred. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** A bid (contract) in 500. */
export interface FiveHundredBidData {
  kind: number;
  tricks: number;
  suit: number;
  value: number;
}

/** A 500 player's per-round state. */
export interface FiveHundredPlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  team: number;
  trickCount: number;
  bid?: FiveHundredBidData | null;
  passed: boolean;
  isDeclarer: boolean;
}

/** A card played in a 500 trick. */
export interface FiveHundredTrickCard {
  playerIdx: number;
  card: Card;
}

/** 500 game configuration. */
export interface FiveHundredConfig {
  cpuDifficulty: number;
  targetScore: number;
}

/** A suggested hint for 500. */
export interface FiveHundredHint {
  bidKind?: number;
  bidTricks?: number;
  bidSuit?: number;
  pass?: boolean;
  discardIndices?: number[];
  cardIndex?: number;
  jokerSuit?: number;
  reason: string;
}

/** Full 500 game state returned from the API. */
export interface FiveHundredResponse extends BaseGameResponse {
  players: FiveHundredPlayerData[];
  phase: number;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  bidPlayerIdx: number;
  dealerIdx: number;
  leadPlayerIdx: number;
  trumpSuit: number;
  contractKind: number;
  contractTricks: number;
  contractValue: number;
  declarerIdx: number;
  highestBid?: FiveHundredBidData | null;
  highestBidder: number;
  jokerLeadSuit: number;
  kittyCount: number;
  currentTrick: FiveHundredTrickCard[];
  teamScores: [number, number];
  gameEndFlag: boolean;
  winnerTeam: number;
  config: FiveHundredConfig;
  hint?: FiveHundredHint;
}

// --- Rook (ルーク) ---
