// Type declarations for mighty. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Mighty player data with bid, roles, scores, and point-card count. */
export interface MightyPlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  bid: number;
  bidNoTrump: boolean;
  isDeclarer: boolean;
  isPartner: boolean;
  partnerRevealed: boolean;
  pointCards: number;
  roundScore: number;
  cumulativeScore: number;
  trickCount: number;
}

/** A card played in a Mighty trick. */
export interface MightyTrickCard {
  playerIdx: number;
  card: Card;
  isJokerLead?: boolean;
  leadDemandSuit?: number;
}

/** Mighty game configuration. */
export interface MightyConfig {
  cpuDifficulty: number;
  minBid: number;
  noTrumpExtra: number;
  pointLimit: number;
}

/** A suggested hint for Mighty. */
export interface MightyHint {
  cardIndex?: number;
  bid?: number;
  bidNoTrump?: boolean;
  trumpSuit?: number;
  partnerSuit?: number;
  partnerValue?: number;
  discardIndices?: number[];
  jokerLeadSuit?: number;
  reason: string;
}

/** Full Mighty game state returned from the API. */
export interface MightyResponse extends BaseGameResponse {
  players: MightyPlayerData[];
  phase: number;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  bidPlayerIdx: number;
  currentTrick: MightyTrickCard[];
  trumpSuit: number;
  partnerCard?: Card | null;
  declarerIdx: number;
  partnerIdx: number;
  partnerRevealed: boolean;
  highestBid: number;
  highestBidder: number;
  winningBidNoTrump: boolean;
  kitty?: Card[];
  gameEndFlag: boolean;
  winnerTeam: number;
  leadPlayerIdx: number;
  config: MightyConfig;
  hint?: MightyHint;
}

// --- 500 (Five Hundred) ---
