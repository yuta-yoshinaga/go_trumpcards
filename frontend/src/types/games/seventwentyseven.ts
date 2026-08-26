// Type declarations for sevenTwentySeven. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** SevenTwentySeven game phase (0=Draw, 1=Result). */
export type SevenTwentySevenPhaseValue = 0 | 1;

/** A SevenTwentySeven player's public/own state. */
export interface SevenTwentySevenPlayer {
  id: number;
  isHuman: boolean;
  /** Remaining chips. */
  chips: number;
  /** Whether this player has said "no more cards" for the round. */
  standing: boolean;
  /** Whether the player has been eliminated (busted) from the match. */
  out: boolean;
  /** Chips this player has wagered into the pot this round. */
  roundBet: number;
  cardCount: number;
  cards: Card[];
  /**
   * The 7-side and 27-side totals, already formatted (`"6.5"`, `"21"`), or
   * `"-"` when that side has gone over. Empty for a player whose cards are
   * hidden — the totals *are* the hand, so hiding one means hiding both.
   *
   * They arrive as strings because the scores move in halves: sending them as
   * JSON numbers would push the rounding decision onto the client.
   */
  lowScore: string;
  highScore: string;
  /** Whether this player took the 7 side / the 27 side. **Both means a scoop.** */
  wonLow: boolean;
  wonHigh: boolean;
}

/** SevenTwentySeven local-rule configuration. */
export interface SevenTwentySevenConfig {
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
 * A suggested hint, computed by the backend.
 *
 * `reason` names **which side it is chasing** — "draw" on its own is not advice
 * when 7 and 27 pull in opposite directions.
 */
export interface SevenTwentySevenHint {
  /** Whether to take another card. `false` means stand pat. */
  draw: boolean;
  /** i18n reason suffix (`chase_seven`, `exactly_twentyseven`, …). */
  reason: string;
}

/**
 * Full SevenTwentySeven game state returned from the API.
 *
 * Every player antes and is dealt 2 cards, then chooses each round to take
 * another card or stand pat, until everybody has stood. **Card values are
 * unusual**: face cards count half a point and an ace counts 1 *or* 11, chosen
 * per hand. At the showdown the hand closest to 7 without going over and the
 * hand closest to 27 without going over each take half the pot; one player
 * taking both scoops it all, and if everybody has gone over both ways the pot
 * carries to the next round.
 */
export interface SevenTwentySevenResponse extends BaseGameResponse {
  players: SevenTwentySevenPlayer[];
  /** Game phase: 0=Draw, 1=Result. */
  phase: SevenTwentySevenPhaseValue;
  roundNumber: number;
  /** Which pass of "card or stand" this is (1-based). */
  drawRound: number;
  /** Chips currently in the pot. */
  pot: number;
  /** Chips carried over from rounds where everybody busted. */
  carryPot: number;
  /** How many rounds in a row the pot has carried. */
  carryCount: number;
  /** Chips each player antes at the start of a round. */
  ante: number;
  /** The human's remaining chip stack. */
  chips: number;
  /** Seat that took the 7 side, or -1. */
  lowWinner: number;
  /** Seat that took the 27 side, or -1. */
  highWinner: number;
  /** Winning seat index of the match, or -1 until it is decided. */
  matchWinnerIdx: number;
  /** The human's round result: 1=won a share, 0=none, -1=took neither side. */
  result: number;
  gameEndFlag: boolean;
  hint?: SevenTwentySevenHint | null;
  config: SevenTwentySevenConfig;
}

// --- Anaconda (Pass the Trash) ---
