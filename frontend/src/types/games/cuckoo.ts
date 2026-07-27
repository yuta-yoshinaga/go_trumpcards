// Type declarations for cuckoo. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Cuckoo phase value: 0=Turn, 1=Refuse, 2=RoundEnd, 3=GameEnd. */
export type CuckooPhaseValue = 0 | 1 | 2 | 3;

/**
 * A single Cuckoo player as returned from the API.
 *
 * Four seats each hold a single card. The human (seat 0) always sees their own
 * card; opponents' cards are `null` until they are revealed at round end or by
 * a King reveal.
 */
export interface CuckooPlayer {
  /** Seat index (0 = human). */
  id: number;
  isHuman: boolean;
  /** The player's single card, or `null` while hidden / eliminated. */
  card: Card | null;
  /** Remaining lives (♥). 0 means eliminated. */
  lives: number;
  /** Whether this player has been knocked out of the game. */
  isEliminated: boolean;
  /** Whether this player has revealed a King to block a swap. */
  kingRevealed: boolean;
  /** Whether it is currently this player's turn. */
  isCurrentTurn: boolean;
}

/**
 * Full Cuckoo (a.k.a. Chase the Ace / Ranter-Go-Round) game state.
 *
 * A simple 4-player life-survival game. Each player holds one card and three
 * lives. On your turn you keep your card or swap it with your neighbour (the
 * dealer swaps with the stock); a King holder may refuse an incoming swap by
 * revealing the King. After everyone acts, the holder(s) of the lowest card
 * lose a life; at zero lives a player is eliminated. The last player standing
 * wins.
 */
export interface CuckooResponse extends BaseGameResponse {
  players: CuckooPlayer[];
  phase: CuckooPhaseValue;
  /** Current round number (1-based). */
  roundNumber: number;
  /** Seat index whose turn it currently is. */
  currentPlayerIdx: number;
  /** Seat index of the dealer this round. */
  dealerIdx: number;
  /** Cards remaining in the stock. */
  stockCount: number;
  gameEndFlag: boolean;
  /** Winning seat index, or -1 until the game ends. */
  winnerIdx: number;
  /** Seat attempting a swap, or -1. */
  pendingSwapFrom: number;
  /** Target seat of a pending swap (the King holder who may refuse), or -1. */
  pendingSwapTo: number;
  /** The lowest card value held this round, or -1 until decided. */
  roundLowest: number;
  /** Seat indices that held the lowest card and lost a life this round. */
  roundLosers: number[];
  config: CuckooConfig;
}

/** Cuckoo configuration as returned from the API. */
export interface CuckooConfig {
  cpuDifficulty: number;
  initialLives: number;
}
