// Type declarations for andarbahar. Split-file layout introduced by issue #4366;
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Which column a card went to, and which one a bet backs. */
export const ANDAR = 0;
/** Bahar — the outer column. */
export const BAHAR = 1;

/**
 * A side bet on how many cards it takes to settle the round.
 *
 * The bands cover every reachable count (1..49) with no gaps.
 */
export const ANDAR_BAHAR_SIDE_NONE = -1;

/** Payout multipliers are stored in tenths so 0.9:1 stays exact in integer chips. */
export const ANDAR_BAHAR_PAYOUT_SCALE = 10;

/** Total returned (stake included) when the first-dealt column wins: 1.9x. */
export const ANDAR_BAHAR_FIRST_COLUMN_PAYOUT = 19;

/** Total returned (stake included) when the second column wins: 2.0x. */
export const ANDAR_BAHAR_SECOND_COLUMN_PAYOUT = 20;

/** One side-bet band: the inclusive card-count range and its payout in tenths. */
export interface AndarBaharSideBand {
  band: number;
  lo: number;
  hi: number;
  /** Total returned per unit staked, in tenths (150 = 15.0x). */
  payout: number;
}

/**
 * The side-bet table, mirroring `andarBaharSideBands` / `andarBaharSidePayouts`
 * in `internal/domain/AndarBahar.go`.
 */
export const ANDAR_BAHAR_SIDE_BANDS: readonly AndarBaharSideBand[] = [
  { band: 0, lo: 1, hi: 1, payout: 150 },
  { band: 1, lo: 2, hi: 5, payout: 42 },
  { band: 2, lo: 6, hi: 10, payout: 41 },
  { band: 3, lo: 11, hi: 15, payout: 52 },
  { band: 4, lo: 16, hi: 25, payout: 41 },
  { band: 5, lo: 26, hi: 35, payout: 90 },
  { band: 6, lo: 36, hi: 51, payout: 330 },
] as const;

/**
 * Response payload for `/andarbahar/exec`.
 *
 * **The asymmetry belongs to `firstColumn`, not to the first card.** The
 * column dealt first gets one extra chance and wins 51.50% of the time, so it
 * pays 0.9:1 while the other pays 1:1.
 */
export interface AndarBaharResponse extends BaseGameResponse {
  /** The reference card. Its colour decides which column is dealt first. */
  joker?: Card;
  /** Cards dealt to Andar. Always an array — the server never sends null. */
  andarCards: Card[];
  /** Cards dealt to Bahar. Always an array — the server never sends null. */
  baharCards: Card[];
  /** The column dealt first, and the one paying 0.9:1. */
  firstColumn: number;
  /** Cards dealt before the round was decided. */
  dealtCount: number;
  /** 1 = Bet, 2 = End. */
  phase: number;
  chips: number;
  betAmount: number;
  betTarget: number;
  sideAmount: number;
  /** -1 when no side bet was placed. */
  sideBand: number;
  /** The column that matched the joker, or -1 until decided. */
  winner: number;
  result: number;
  /** Total returned to the player, stake included. */
  payout: number;
  /** Winning column per round. Always an array. */
  history: number[];
}
