// Type declarations for napoleon. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Napoleon player data with bid, roles, scores, and picture card count. */
export interface NapoleonPlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  bid: number;
  isNapoleon: boolean;
  isAdjutant: boolean;
  adjutantRevealed: boolean;
  pictureCards: number;
  roundScore: number;
  cumulativeScore: number;
  trickCount: number;
}

/** A card played in a Napoleon trick. */
export interface NapoleonTrickCard {
  playerIdx: number;
  card: Card;
}

/** Napoleon game configuration. */
export interface NapoleonConfig {
  cpuDifficulty: number;
  minBid: number;
  pointLimit: number;
}

/** A suggested hint for Napoleon. */
export interface NapoleonHint {
  cardIndex?: number;
  bid?: number;
  trumpSuit?: number;
  adjutantSuit?: number;
  adjutantValue?: number;
  discardIndex?: number;
  reason: string;
}

/** Full Napoleon game state returned from the API. */
export interface NapoleonResponse extends BaseGameResponse {
  players: NapoleonPlayerData[];
  phase: number;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  bidPlayerIdx: number;
  currentTrick: NapoleonTrickCard[];
  trumpSuit: number;
  adjutantCard: Card | null;
  napoleonIdx: number;
  adjutantIdx: number;
  adjutantRevealed: boolean;
  highestBid: number;
  highestBidder: number;
  kitty: Card[];
  gameEndFlag: boolean;
  winnerTeam: number;
  config: NapoleonConfig;
  hint?: NapoleonHint;
}

// --- Mighty (マイティ) ---
