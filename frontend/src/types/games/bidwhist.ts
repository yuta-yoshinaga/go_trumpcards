// Type declarations for bidwhist. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** A Bid Whist bid: target tricks over the book plus a direction. */
export interface BidWhistBidData {
  tricks: number;
  direction: number;
}

/** A Bid Whist player's per-round state. */
export interface BidWhistPlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  team: number;
  trickCount: number;
  bid?: BidWhistBidData | null;
  passed: boolean;
  isDeclarer: boolean;
}

/** A card played in a Bid Whist trick. */
export interface BidWhistTrickCard {
  playerIdx: number;
  card: Card;
}

/** Bid Whist game configuration. */
export interface BidWhistConfig {
  cpuDifficulty: number;
  targetScore: number;
}

/** A suggested hint for Bid Whist. */
export interface BidWhistHint {
  bidTricks?: number;
  bidDirection?: number;
  pass?: boolean;
  trumpSuit?: number;
  discardIndices?: number[];
  cardIndex?: number;
  reason: string;
}

/** Full Bid Whist game state returned from the API. */
export interface BidWhistResponse extends BaseGameResponse {
  players: BidWhistPlayerData[];
  phase: number;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  bidPlayerIdx: number;
  dealerIdx: number;
  leadPlayerIdx: number;
  trumpSuit: number;
  contractTricks: number;
  contractDirection: number;
  declarerIdx: number;
  highestBid?: BidWhistBidData | null;
  highestBidder: number;
  kittyCount: number;
  /**
   * Indices, within the human declarer's hand, of the six cards that came from
   * the kitty. Populated only while the human is exchanging during
   * KITTY_EXCHANGE; empty in every other phase and for a CPU declarer.
   */
  kittyIndices: number[];
  currentTrick: BidWhistTrickCard[];
  teamScores: [number, number];
  gameEndFlag: boolean;
  winnerTeam: number;
  config: BidWhistConfig;
  hint?: BidWhistHint;
}

// --- Skat (スカート) ---
