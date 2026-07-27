// Type declarations for rook. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** A Rook player's per-round state. */
export interface RookPlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  team: number;
  trickCount: number;
  points: number;
  bid: number;
  passed: boolean;
  isDeclarer: boolean;
}

/** A card played in a Rook trick. */
export interface RookTrickCard {
  playerIdx: number;
  card: Card;
}

/** Rook game configuration. */
export interface RookConfig {
  cpuDifficulty: number;
  targetScore: number;
}

/** A suggested hint for Rook. */
export interface RookHint {
  bid?: number;
  pass?: boolean;
  discardIndices?: number[];
  trumpColor?: number;
  cardIndex?: number;
  reason: string;
}

/** Full Rook game state returned from the API. */
export interface RookResponse extends BaseGameResponse {
  players: RookPlayerData[];
  phase: number;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  bidPlayerIdx: number;
  dealerIdx: number;
  leadPlayerIdx: number;
  /** Declared trump color (1=red, 2=gold, 3=green, 4=black; -1 until declared). */
  trumpColor: number;
  contractBid: number;
  declarerIdx: number;
  highestBid: number;
  highestBidder: number;
  nestCount: number;
  /** Nest (widow) cards; only populated for the human declarer during nest exchange. */
  nest: Card[];
  currentTrick: RookTrickCard[];
  teamScores: [number, number];
  teamPoints: [number, number];
  gameEndFlag: boolean;
  winnerTeam: number;
  config: RookConfig;
  hint?: RookHint;
}

// --- Bid Whist (ビッド・ホイスト) ---
