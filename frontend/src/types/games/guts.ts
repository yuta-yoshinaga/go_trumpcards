// Type declarations for guts. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Guts game phase (0=Declare, 1=Result). */
export type GutsPhaseValue = 0 | 1;

/**
 * A Guts player's public/own state. `cards` is populated for the human and for
 * every player still `in` at showdown; `handName` is an i18n suffix
 * (`"pair"`, `"highcard"`, or `""`) set only when a hand is revealed.
 */
export interface GutsPlayer {
  id: number;
  isHuman: boolean;
  /** Remaining chips. */
  chips: number;
  /** Whether the player declared in (stayed) this round. */
  in: boolean;
  /** Whether the player has been eliminated (busted) from the match. */
  out: boolean;
  /** Chips this player has wagered / owes into the pot this round. */
  roundBet: number;
  cardCount: number;
  cards: Card[];
  /** The revealed hand-rank i18n suffix (`"pair"`, `"highcard"`), or empty. */
  handName?: string;
  /** Whether this player won the round's pot. */
  isWinner: boolean;
  /** Whether this player stayed in but lost and must match the pot. */
  isMatcher: boolean;
}

/** Guts local-rule configuration. */
export interface GutsConfig {
  /** Number of players at the table (2–7). */
  playerCount: number;
  /** Chips each player antes into the pot at the start of a round. */
  ante: number;
  /** Chips each player begins the match with. */
  startingChips: number;
  /** Number of rounds after which the richest player wins the match. */
  targetRounds: number;
}

/**
 * A suggested hint for Guts, computed by the backend. `declaration` is the
 * suggested call (0=out, 1=in) and `reason` is an i18n reason suffix
 * (`strong_hand` / `weak_hand`).
 */
export interface GutsHint {
  /** Suggested declaration: 0=out (fold), 1=in (stay). */
  declaration: number;
  /** i18n reason suffix identifier. */
  reason: string;
}

/**
 * Full Guts game state returned from the API.
 *
 * Guts is a fast multi-player pot-vying gambling game: every player antes, is
 * dealt 2 cards, and simultaneously declares "in" (stay) or "out" (fold). The
 * players who stayed reveal their hands; the best hand wins the pot, and every
 * other player who stayed must match the pot for the next round (which
 * carries over). If nobody stays, the pot carries to the next round. When
 * `targetRounds` rounds have been played the richest player wins the match.
 */
export interface GutsResponse extends BaseGameResponse {
  players: GutsPlayer[];
  /** Game phase: 0=Declare, 1=Result. */
  phase: GutsPhaseValue;
  roundNumber: number;
  /** Chips currently in the pot. */
  pot: number;
  /** Chips carried over from rounds where nobody stayed. */
  carryPot: number;
  /** Chips each player antes at the start of a round. */
  ante: number;
  /** The human's remaining chip stack. */
  chips: number;
  /** Winning seat index of the current round, or -1 for none. */
  winnerIdx: number;
  /** Winning seat index of the match, or -1 until it is decided. */
  matchWinnerIdx: number;
  /** The human's round result: 1=win, 0=none, -1=lose (matched). */
  result: number;
  /** Seat indices of players who stayed in but lost and must match the pot. */
  matchers: number[];
  gameEndFlag: boolean;
  hint?: GutsHint | null;
  config: GutsConfig;
}

// --- Anaconda (Pass the Trash) ---
