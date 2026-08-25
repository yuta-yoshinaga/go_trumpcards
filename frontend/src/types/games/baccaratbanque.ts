// Type declarations for baccaratbanque. Split-file layout introduced by issue #4366;
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Seats at a Baccarat Banque table: one bank and two tableaux. */
export const BACCARAT_BANQUE_SEATS = 3;

/** The banker's seat. The human always holds the bank. */
export const BACCARAT_BANQUE_BANKER_SEAT = 0;

/** The right tableau. */
export const BACCARAT_BANQUE_RIGHT_SEAT = 1;

/** The left tableau. */
export const BACCARAT_BANQUE_LEFT_SEAT = 2;

/**
 * Phases, matching the Go domain.
 *
 * **These are strings, not indices.** The punters resolve inside the deal, so
 * the first phase a human ever sees is `banker`.
 */
export const BACCARAT_BANQUE_PHASE = {
  PUNTERS: 'punters',
  BANKER: 'banker',
  RESULT: 'result',
  GAME_END: 'gameEnd',
} as const;

/** One of the three seats. */
export type BaccaratBanqueRole = 'banker' | 'right' | 'left';

/** How one tableau finished against the bank. */
export type BaccaratBanqueOutcome = 'bankerWin' | 'punterWin' | 'tie';

/** One seat's cards and chips. **Baccarat deals face up**, so `cards` is never hidden. */
export interface BaccaratBanquePlayer {
  id: number;
  isHuman: boolean;
  role: BaccaratBanqueRole;
  cards: Card[];
  total: number;
  chips: number;
  /** What this tableau staked. `0` for the banker, who covers both. */
  bet: number;
  /** Whether this seat took a third card in the current coup. */
  drawn: boolean;
}

/**
 * How the bank finished against one tableau.
 *
 * **The two sides settle separately**, so one coup can pay the right and
 * collect from the left; `delta` is that side's own movement.
 */
export interface BaccaratBanqueSide {
  seatIdx: number;
  outcome: BaccaratBanqueOutcome;
  bet: number;
  delta: number;
}

/** One coup's settlement. */
export interface BaccaratBanqueResult {
  bankerTotal: number;
  sides: BaccaratBanqueSide[];
  /** The bank's net across both tableaux — the two deltas summed. */
  bankerDelta: number;
  bankerNatural: boolean;
}

/** Baccarat Banque game settings. */
export interface BaccaratBanqueConfig {
  cpuDifficulty: number;
  startChips: number;
  betAmount: number;
}

/** Response payload for `/baccaratbanque/exec`. */
export interface BaccaratBanqueResponse extends BaseGameResponse {
  players: BaccaratBanquePlayer[];
  /** `punters` | `banker` | `result` | `gameEnd`. */
  phase: string;
  coupNumber: number;
  /**
   * Coups held on this bank.
   *
   * **A loss does not reset it.** The bank stays with the same seat until the
   * shoe runs out, the banker retires, or the bank is broken — which is the
   * one rule separating this game from Chemin de Fer.
   */
  bankHeld: number;
  /** Cards left in the three-pack shoe. The bank ends when a coup will not fit. */
  shoeRemaining: number;
  /** Whether the banker gave the bank up rather than being broken or running out. */
  retired: boolean;
  lastResult?: BaccaratBanqueResult;
  gameEndFlag: boolean;
  winnerIdx: number;
  isHumanTurn: boolean;
  /** Whether the hint recommends taking a third card. */
  hintDraw: boolean;
  /** `low_total`, `behind_both`, `stand`, `natural` or `none`. */
  hintReason: string;
  config?: BaccaratBanqueConfig;
}
