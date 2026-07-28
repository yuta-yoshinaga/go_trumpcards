// Type declarations for bouillotte. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Bouillotte game phase (0=Betting, 1=Result). */
export type BouillottePhaseValue = 0 | 1;

/**
 * A Bouillotte player's public/own state. `cards` is populated for the human
 * and, at the result phase, for every player who has not folded; `handName` is
 * an i18n suffix (`"brelan"`, `"highcard"`, or `""`) set only when a hand is
 * revealed.
 */
export interface BouillottePlayer {
  id: number;
  isHuman: boolean;
  /** Remaining chips. */
  chips: number;
  /** Chips this player has wagered into the pot this round. */
  roundBet: number;
  /** Whether the player has folded out of the current round. */
  folded: boolean;
  /** Whether the player has been eliminated (busted) from the match. */
  out: boolean;
  cardCount: number;
  cards: Card[];
  /** The revealed hand-rank i18n suffix (`"brelan"`, `"highcard"`), or empty. */
  handName?: string;
  /** Whether this player won the round's pot. */
  isWinner: boolean;
}

/** Bouillotte local-rule configuration. */
export interface BouillotteConfig {
  /** Number of players at the table (3–4). */
  playerCount: number;
  /** Chips each player antes into the pot at the start of a round. */
  ante: number;
  /** Chips each player begins the match with. */
  startingChips: number;
  /** Number of rounds after which the richest player wins the match. */
  targetRounds: number;
}

/**
 * A suggested hint for Bouillotte, computed by the backend. `action` is the
 * suggested betting action (`"call"` / `"raise"` / `"fold"`) and `reason` is an
 * i18n reason suffix (`strong_hand` / `medium_hand` / `weak_hand`).
 */
export interface BouillotteHint {
  /** Suggested betting action: `"call"`, `"raise"`, or `"fold"`. */
  action: string;
  /** i18n reason suffix identifier. */
  reason: string;
}

/**
 * Full Bouillotte game state returned from the API.
 *
 * Bouillotte is an 18th-century French 3-card poker-vying pot game. Each player
 * antes, is dealt 3 cards, and a shared "retourne" card is turned up. Players
 * take turns to call, raise (vie), or fold; when betting closes the non-folded
 * players reveal their hands and the best hand (a brelan of three matching
 * ranks — counting the retourne — beats a high card) takes the pot. Chips
 * accumulate across rounds; after `targetRounds` rounds the richest player wins.
 */
export interface BouillotteResponse extends BaseGameResponse {
  players: BouillottePlayer[];
  /** Game phase: 0=Betting, 1=Result. */
  phase: BouillottePhaseValue;
  roundNumber: number;
  /** Chips currently in the pot. */
  pot: number;
  /** Chips each player antes at the start of a round. */
  ante: number;
  /** The human's remaining chip stack. */
  chips: number;
  /** The current bet each active player must match to stay in. */
  currentBet: number;
  /** Number of raises made this round. */
  raiseCount: number;
  /** Maximum raises permitted this round. */
  maxRaises: number;
  /** Seat index of the player to act. */
  currentPlayerIdx: number;
  /** Seat index of the dealer. */
  dealerIdx: number;
  /** The shared turned-up "retourne" card, or null before it is dealt. */
  retourne: Card | null;
  /** Whether it is the human's turn to act. */
  isHumanTurn: boolean;
  /** Whether the human may currently raise (vie). */
  canRaise: boolean;
  /** Winning seat index of the current round, or -1 for none. */
  winnerIdx: number;
  /** Winning seat index of the match, or -1 until it is decided. */
  matchWinnerIdx: number;
  /** The human's round result: 1=win, 0=none, -1=lose. */
  result: number;
  gameEndFlag: boolean;
  hint?: BouillotteHint | null;
  config: BouillotteConfig;
}
