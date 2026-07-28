// Type declarations for openfacechinese. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** A single player as returned from the Open Face Chinese Poker (OFC) API. */
export interface OpenFaceChinesePlayer {
  /** Seat index (0 = human). */
  id: number;
  /** True for the human player. */
  isHuman: boolean;
  /** Top row (up to 3 cards). */
  front: Card[];
  /** Middle row (up to 5 cards). */
  middle: Card[];
  /** Bottom row (up to 5 cards). */
  back: Card[];
  /** The pending card(s) awaiting placement (human only; empty for CPU). */
  pending: Card[];
  /** Net points scored in the just-finished round. */
  roundScore: number;
  /** Royalty bonus points earned this round. */
  royalty: number;
  /** True when the hand is fouled (rows not in non-decreasing strength order). */
  fouled: boolean;
  /** True when the player qualified for Fantasyland. */
  fantasyland: boolean;
  /** Cumulative score across all rounds. */
  totalScore: number;
}

/** Open Face Chinese Poker (OFC) config echoed back by the server. */
export interface OpenFaceChineseConfig {
  cpuDifficulty: number;
  playerCount: number;
  targetRounds: number;
}

/** A placement hint returned by the Open Face Chinese Poker (OFC) /openfacechinese/exec endpoint. */
export interface OpenFaceChineseHint {
  /** Suggested row (0=front, 1=middle, 2=back). */
  row: number;
  /** Human-readable rationale for the suggestion. */
  reason: string;
}

/** Server response for the Open Face Chinese Poker (OFC) game (POST /openfacechinese/exec). */
export interface OpenFaceChineseResponse extends BaseGameResponse {
  /** Current phase (0=Placing, 1=RoundEnd, 2=GameEnd). */
  phase: number;
  /** 1-based round number. */
  roundNumber: number;
  /** Seat index whose turn it currently is. */
  currentPlayerIdx: number;
  /** Seat index of the current dealer. */
  dealerIdx: number;
  /** The card the human must place this turn (present only on the human's turn). */
  currentCard?: Card;
  /** True when the game has ended. */
  gameEndFlag: boolean;
  /** Winning seat index, or -1 for a draw. */
  winnerIdx: number;
  /** True when it is the human player's turn to place a card. */
  isHumanTurn: boolean;
  /** Optional placement hint (present only on a hint request). */
  hint?: OpenFaceChineseHint;
  /** One entry per player. */
  players: OpenFaceChinesePlayer[];
  /** Echoed game configuration. */
  config: OpenFaceChineseConfig;
}
