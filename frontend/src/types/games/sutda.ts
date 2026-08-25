// Type declarations for sutda. Follows the split-out convention of card.ts
// (issue #4366); card.ts re-exports this file.

import type { BaseGameResponse, Card } from '../common';

/** One seat at the table. */
export interface SutdaPlayer {
  id: number;
  isHuman: boolean;
  cardCount: number;
  /**
   * The cards this seat shows.
   *
   * **Empty until the showdown** for everyone but you — that is what makes the
   * betting worth anything.
   */
  cards: Card[];
  chips: number;
  /** What this seat has put in this hand. */
  bet: number;
  folded: boolean;
  revealed: boolean;
  /** Stable identifier for the hand, for i18n. Empty while the cards are down. */
  handName: string;
  /** Strength of the hand. 0 while the cards are down. */
  handRank: number;
  isDealer: boolean;
}

/** One hand's result. */
export interface SutdaResult {
  winners: number[];
  pot: number;
  /** Stable hand identifiers, indexed by seat. */
  handNames: string[];
  folded: boolean[];
}

/** Sutda game configuration. */
export interface SutdaConfig {
  cpuDifficulty: number;
  /** Seats at the table, 2-5. */
  seats: number;
  startChips: number;
}

/**
 * Full Sutda game state returned from the API.
 *
 * A Korean betting game on a **20-card hanafuda pack** (months 1-10, two of
 * each) where every hand is exactly two cards. **Only months 1, 3 and 8 carry a
 * gwang (bright)**, which is why there are exactly three gwang-ttaeng.
 */
export interface SutdaResponse extends BaseGameResponse {
  players: SutdaPlayer[];
  /** "bet" | "showdown" | "gameEnd". */
  phase: string;
  handNumber: number;
  dealerIdx: number;
  currentPlayerIdx: number;
  pot: number;
  currentBet: number;
  /** What you owe to call. 0 means you may check. */
  callAmount: number;
  /** Already accounts for the raise cap and your chips. */
  canRaise: boolean;
  raiseCount: number;
  maxRaises: number;
  betUnit: number;
  /** Your own hand's identifier — always visible. */
  humanHandName: string;
  lastResult?: SutdaResult | null;
  gameEndFlag: boolean;
  winnerIdx: number;
  isHumanTurn: boolean;
  isShowdown: boolean;
  /** Action the backend suggests ("call" | "raise" | "fold"), or "". */
  hintAction: string;
  /** i18n reason identifier for the suggestion. */
  hintReason: string;
  config: SutdaConfig;
}
