// Type declarations for costlycolours. Follows the split-out convention of
// card.ts (issue #4366); card.ts re-exports this file.

import type { BaseGameResponse, Card } from '../common';

/** One seat at the two-handed table. */
export interface CostlyColoursPlayer {
  id: number;
  isHuman: boolean;
  /** Hand cards. Populated only for the human. */
  cards: Card[];
  cardCount: number;
  /** Cards played this deal. The show counts these plus what is still held. */
  played: Card[];
  /** Running match score. */
  score: number;
  isDealer: boolean;
  /** Whether this seat took a card in the exchange. */
  moggedIn: boolean;
}

/** One scoring category of the show. */
export interface CostlyColoursScoreLine {
  /** "jackDeuce" | "rank" | "colour". */
  key: string;
  /** Points per seat. */
  points: number[];
}

/** One deal's scoring result. */
export interface CostlyColoursResult {
  lines: CostlyColoursScoreLine[];
  totals: number[];
  /** The colour combination each seat scored, or "" for none. */
  combos: string[];
}

/** Costly Colours game configuration. */
export interface CostlyColoursConfig {
  cpuDifficulty: number;
  /** Points needed to win, 31-121 (61 = Cotton, 121 = Parlett). */
  targetScore: number;
}

/**
 * Full Costly Colours game state returned from the API.
 *
 * A Shropshire ancestor of Cribbage. **The count scores at 15, 25 and 31** —
 * the 25 has no counterpart in Cribbage — and the show counts a ladder of
 * colour combinations over the three cards plus the turn-up, topped by four in
 * suit: "Costly Colours".
 */
export interface CostlyColoursResponse extends BaseGameResponse {
  players: CostlyColoursPlayer[];
  /** "mog" | "play" | "show" | "gameEnd". */
  phase: string;
  dealNumber: number;
  dealerIdx: number;
  currentPlayerIdx: number;
  /** The card turned for trumps. Counted in the show. */
  turnUp?: Card | null;
  /** Cards played in the current count. */
  pile: Card[];
  /** Running count. Never above 31. */
  total: number;
  /** Seat that said "go", or -1. */
  wentOut: number;
  /** Hand indices that fit under 31. Empty means go, or not the human's turn. */
  playableIdxs: number[];
  lastResult?: CostlyColoursResult | null;
  gameEndFlag: boolean;
  winnerIdx: number;
  isHumanTurn: boolean;
  /** Recommended hand card, or -1 (the mog phase names no card). */
  hintHandIdx: number;
  /** Whether the hint says to accept the exchange. */
  hintAcceptMog: boolean;
  hintReason: string;
  config: CostlyColoursConfig;
}
