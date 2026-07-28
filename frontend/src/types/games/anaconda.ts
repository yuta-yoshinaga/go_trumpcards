// Type declarations for anaconda. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Anaconda game phase (0=Pass, 1=Set, 2=Roll, 3=Result). */
export type AnacondaPhaseValue = 0 | 1 | 2 | 3;

/**
 * An Anaconda player's public/own state. `cards` holds the revealed cards: the
 * human sees their full hand, CPUs show only the `rollIndex`-length prefix
 * revealed so far during Roll (and the full 5 at showdown if still active).
 * `handName` is a poker-category i18n suffix (`"fourkind"`, `"flush"`, …) set
 * only when a full 5-card hand is revealed.
 */
export interface AnacondaPlayer {
  id: number;
  isHuman: boolean;
  /** Remaining chips. */
  chips: number;
  /** Whether the player folded out of the current round. */
  folded: boolean;
  /** Whether the player has been eliminated (busted) from the match. */
  out: boolean;
  /** Chips this player has wagered into the pot across the whole round. */
  roundBet: number;
  /** Chips this player has wagered on the current betting street. */
  streetBet: number;
  cardCount: number;
  /** Revealed cards (see interface doc). */
  cards: Card[];
  /** The revealed poker-category i18n suffix, or empty until a 5-card hand shows. */
  handName?: string;
  /** Whether this player won the round's pot. */
  isWinner: boolean;
}

/** Anaconda local-rule configuration. */
export interface AnacondaConfig {
  /** Number of players at the table (3–7). */
  playerCount: number;
  /** Chips each player antes into the pot at the start of a round. */
  ante: number;
  /** Chips each player begins the match with. */
  startingChips: number;
  /** Number of rounds after which the richest player wins the match. */
  targetRounds: number;
}

/**
 * A suggested hint for Anaconda, computed by the backend. `action` is the
 * suggested move (`"pass"` / `"keep"` / `"raise"` / `"call"` / `"fold"`),
 * `cardIndices` accompanies pass/keep suggestions, and `reason` is an i18n
 * reason suffix (`pass_weakest` / `keep_best` / `strong_hand` / `medium_hand` /
 * `weak_hand`).
 */
export interface AnacondaHint {
  /** Suggested action: pass/keep/raise/call/fold. */
  action: string;
  /** Card indices to pass or keep, when the suggestion is a pass/keep. */
  cardIndices?: number[];
  /** i18n reason suffix identifier. */
  reason: string;
}

/**
 * Full Anaconda (Pass the Trash) game state returned from the API.
 *
 * Anaconda is a poker pot game: each player is dealt 7 cards, then passes 3,
 * then 2, then 1 card to the left; keeps their best 5 (discarding 2); and
 * reveals them one at a time ("roll") with a betting round between reveals.
 * The best 5-card poker hand among the players still active wins the pot.
 */
export interface AnacondaResponse extends BaseGameResponse {
  players: AnacondaPlayer[];
  /** Game phase: 0=Pass, 1=Set, 2=Roll, 3=Result. */
  phase: AnacondaPhaseValue;
  roundNumber: number;
  /** Seat index of the current dealer. */
  dealerIdx: number;
  /** Seat index of the player to act. */
  currentPlayer: number;
  /** Cards still to pass this sub-round (3/2/1 during Pass, 0 otherwise). */
  passCount: number;
  /** Cards revealed so far during Roll (0–5). */
  rollIndex: number;
  /** Chips currently in the pot. */
  pot: number;
  /** The current bet to match on this betting street. */
  currentBet: number;
  /** Number of raises already made on the current street. */
  raiseCount: number;
  /** The maximum number of raises allowed per street. */
  maxRaises: number;
  /** Chips each player antes at the start of a round. */
  ante: number;
  /** The human's remaining chip stack. */
  chips: number;
  /** Winning seat index of the current round, or -1 for none. */
  winnerIdx: number;
  /** Winning seat index of the match, or -1 until it is decided. */
  matchWinnerIdx: number;
  /** The human's round result: 1=win, 0=none, -1=lose. */
  result: number;
  gameEndFlag: boolean;
  /** Whether it is the human's turn to act. */
  isHumanTurn: boolean;
  /** Whether the human may raise (raises remain and chips suffice). */
  canRaise: boolean;
  hint?: AnacondaHint | null;
  config: AnacondaConfig;
}
