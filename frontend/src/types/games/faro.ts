// Type declarations for faro. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** A single chip bet placed on one rank of the Faro layout. */
export interface FaroBet {
  /** Rank the chip is placed on (1=A .. 13=K). */
  rank: number;
  /** Wagered chip amount. */
  amount: number;
  /** True when the bet is a "copper" — wagering the rank to lose rather than win. */
  copper: boolean;
}

/** Server response for the Faro game (POST /faro/exec). */
export interface FaroResponse extends BaseGameResponse {
  /** Current phase (1=Betting, 2=Turn, 3=Call, 4=RoundEnd, 5=GameEnd). */
  phase: number;
  /** Player's remaining bankroll in chips. */
  chips: number;
  /** Chips currently placed on the layout, one entry per wagered rank. */
  bets: FaroBet[];
  /** The burned "soda" card (first card of the deal), or null before the deal. */
  soda: Card | null;
  /** The most recent turn's losing card (1st turned — bank collects), or null. */
  losingCard: Card | null;
  /** The most recent turn's winning card (2nd turned — pays the player), or null. */
  winningCard: Card | null;
  /** True when the last turn was a split (both cards the same rank — bank takes half). */
  split: boolean;
  /** Number of turns dealt so far this round. */
  turnsPlayed: number;
  /** Total number of turns in a full round. */
  turnsTotal: number;
  /** Number of cards still left in the dealing box. */
  remaining: number;
  /** The final three cards available to call (populated during the Call phase). */
  callCards: Card[];
  /** The player's predicted order for the called cards (rank values), empty when none. */
  callOrder: number[];
  /** True when the most recent call was correct (paid 4:1). */
  callWon: boolean;
  /** Net chip change for the just-finished round. */
  totalPayout: number;
  gameEndFlag: boolean;
}
