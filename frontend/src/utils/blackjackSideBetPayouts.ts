/**
 * Blackjack side-bet payout multipliers, re-exported from the golden fixture.
 *
 * **Source of truth**: `internal/domain/BlackJackSideBet.go`
 * (`BJPPMixedPairPayout`, `BJPPColoredPairPayout`, `BJPPPerfectPairPayout`,
 * `BJT3FlushPayout`, `BJT3StraightPayout`, `BJT3ThreeOfAKindPayout`,
 * `BJT3StraightFlushPayout`, `BJT3SuitedTripsPayout`).
 *
 * `TestBlackJackSideBetPayouts_GoldenValues` (Go side) reads
 * `frontend/src/constants/blackjackSideBetPayouts.json` and
 * asserts that every value matches the Go constants independently.
 * Changing only one side will fail that side's test, and regenerating the
 * fixture to fix it will break the other side.
 */

import raw from '../constants/blackjackSideBetPayouts.json';

/** Perfect Pairs side-bet payout multipliers. */
export interface PerfectPairsPayout {
  /** Mixed Pair (different colour, same rank) */
  mixed: number;
  /** Coloured Pair (same colour, different suit, same rank) */
  colored: number;
  /** Perfect Pair (same suit, same rank) */
  perfect: number;
}

/** 21+3 (Poker Hand Bonus) side-bet payout multipliers. */
export interface TwentyOnePlus3Payout {
  /** Three cards of the same suit (non-consecutive, non-matching rank) */
  flush: number;
  /** Three consecutive ranks of mixed suits */
  straight: number;
  /** Three of a Kind (same rank, different suits) */
  trips: number;
  /** Straight Flush (consecutive ranks, same suit) */
  straightFlush: number;
  /** Suited Trips (same rank, same suit) */
  suitedTrips: number;
}

/** All blackjack side-bet payout multipliers. */
export interface BlackjackSideBetPayouts {
  perfectPairs: PerfectPairsPayout;
  twentyOnePlus3: TwentyOnePlus3Payout;
}

/**
 * Typed payout multipliers for blackjack side bets.
 * Do **not** hardcode these values elsewhere — update the golden fixture and
 * the Go constants in sync (the cross-side golden test will catch drift).
 */
export const BLACKJACK_SIDE_BET_PAYOUTS: BlackjackSideBetPayouts = raw;
