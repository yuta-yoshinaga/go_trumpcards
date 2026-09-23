// Type declarations for batak. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Batak player data with bid and raw integer scores. */
export interface BatakPlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  /**
   * Player's bid value:
   * - -1: Uncalled / not yet bid
   * - 0: Pass
   * - 5..13: Declared bid
   */
  bid: number;
  /** Round score (raw integer score). */
  roundScore: number;
  /** Cumulative score (raw integer score). */
  cumulativeScore: number;
  trickCount: number;
}

/** A card played in a Batak trick. */
export interface BatakTrickCard {
  playerIdx: number;
  card: Card;
}

/** Batak game configuration. */
export interface BatakConfig {
  cpuDifficulty: number;
  maxRounds: number;
}

/** A suggested hint for Batak. */
export interface BatakHint {
  cardIndex?: number;
  bid?: number;
  reason: string;
}

/** Full Batak game state returned from the API. */
export interface BatakResponse extends BaseGameResponse {
  players: BatakPlayerData[];
  phase: number;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  bidPlayerIdx: number;
  /** Declarer player index (-1 if not yet determined). */
  declarerIdx: number;
  /** Current highest bid (0 if not yet declared). */
  highBid: number;
  /** Minimum legal bid for human player (5..13, 0 if only pass is possible or not human turn). */
  minLegalBid: number;
  currentTrick: BatakTrickCard[];
  spadesBroken: boolean;
  gameEndFlag: boolean;
  winnerIdx: number;
  leadPlayerIdx: number;
  config: BatakConfig;
  hint?: BatakHint;
  /**
   * Indices in the human player's hand that are legal to play this turn.
   * Empty array outside the play phase / when it is not the human's turn.
   */
  validPlayIndices: number[];
}
