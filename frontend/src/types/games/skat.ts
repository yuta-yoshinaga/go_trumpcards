// Type declarations for skat. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** A Skat player's per-round state. */
export interface SkatPlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  bid: number;
  isDeclarer: boolean;
  cardPoints: number;
  roundsWon: number;
  roundsLost: number;
  roundScore: number;
  cumulativeScore: number;
  trickCount: number;
}

/** A card played in a Skat trick. */
export interface SkatTrickCard {
  playerIdx: number;
  card: Card;
}

/** Skat game configuration. */
export interface SkatConfig {
  cpuDifficulty: number;
  targetScore: number;
}

/** A suggested hint for Skat. */
export interface SkatHint {
  cardIndex?: number;
  bid?: number;
  gameType?: number;
  trumpSuit?: number;
  pickSkat?: boolean;
  discardIndex?: number;
  reason: string;
}

/** Full Skat game state returned from the API. */
export interface SkatResponse extends BaseGameResponse {
  players: SkatPlayerData[];
  phase: number;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  currentTrick: SkatTrickCard[];
  forehandIdx: number;
  middlehandIdx: number;
  rearhandIdx: number;
  dealerIdx: number;
  declarerIdx: number;
  currentBid: number;
  activeBidActorIdx: number;
  gameType: number;
  trumpSuit: number;
  skat?: Card[];
  originalSkat?: Card[];
  pickedSkat: boolean;
  declarerCardPoints: number;
  defendersCardPoints: number;
  winnerSide: number;
  gameValue: number;
  gameEndFlag: boolean;
  leadPlayerIdx: number;
  config: SkatConfig;
  hint?: SkatHint;
}

// --- Shithead / Karma (シットヘッド / カーマ) ---
