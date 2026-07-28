// Type declarations for ohhell. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Oh Hell player data with scores. */
export interface OhHellPlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  bid: number;
  roundScore: number;
  cumulativeScore: number;
  trickCount: number;
}

/** A card played in an Oh Hell trick. */
export interface OhHellTrickCard {
  playerIdx: number;
  card: Card;
}

/** Oh Hell game configuration. */
export interface OhHellConfig {
  cpuDifficulty: number;
  maxHandSize: number;
  scoringVariant: number;
  roundDirection: number;
}

/** A suggested hint for Oh Hell. */
export interface OhHellHint {
  cardIndex?: number;
  bid?: number;
  reason: string;
}

/** Full Oh Hell game state returned from the API. */
export interface OhHellResponse extends BaseGameResponse {
  players: OhHellPlayerData[];
  phase: number;
  roundNumber: number;
  totalRounds: number;
  handSize: number;
  trickNumber: number;
  currentPlayerIdx: number;
  bidPlayerIdx: number;
  dealerIdx: number;
  currentTrick: OhHellTrickCard[];
  trumpCard: Card | null;
  trumpSuit: number;
  restrictedBid: number;
  gameEndFlag: boolean;
  winnerIdx: number;
  leadPlayerIdx: number;
  hint?: OhHellHint;
  config: OhHellConfig;
}

// --- Wizard (ウィザード) ---
