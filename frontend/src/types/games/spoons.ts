// Type declarations for spoons. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Spoons game phase (0=Pass 1=Grab 2=RoundEnd 3=GameEnd). */
export type SpoonsPhaseValue = 0 | 1 | 2 | 3;

/**
 * A Spoons player's public/own state. `hand` is non-empty only for the human
 * (seat 0); CPU hands are returned as an empty array.
 */
export interface SpoonsPlayer {
  /** Display name ("あなた" / "CPU"). */
  name: string;
  isHuman: boolean;
  /** Number of cards currently held. */
  handSize: number;
  /** The player's cards — populated only for the human. */
  hand: Card[];
  /** Number of S-P-O-O-N-S letters collected (0–6). */
  letters: number;
  /** Whether the player has been eliminated (6 letters). */
  eliminated: boolean;
  /** Whether the player currently holds a grabbed spoon. */
  hasSpoon: boolean;
}

/**
 * Full Spoons game state returned from the API.
 *
 * Spoons is a 4-player pass-and-grab speed game played with a 52-card deck (4
 * cards each). Players continuously pass a card to the next player; when someone
 * collects four of a kind they grab a spoon and everyone races for the
 * remaining spoons (one fewer than the number of players). The player left
 * without a spoon gains a letter — S, P, O, O, N, S. After six letters that
 * player is eliminated; the last player standing wins.
 */
export interface SpoonsResponse extends BaseGameResponse {
  phase: SpoonsPhaseValue;
  gameEndFlag: boolean;
  /** Winning seat index, or -1 until the game ends. */
  winnerIdx: number;
  /** Seat index whose turn it currently is. */
  currentPlayerIdx: number;
  /** Seat index of the "feeder" who draws from the draw pile this round. */
  feederIdx: number;
  /** Whether it is currently the human's turn to act. */
  isHumanTurn: boolean;
  /** Spoons still on the table to be grabbed. */
  spoonsRemaining: number;
  /** Whether the grab window is open (race to grab a spoon). */
  grabWindowOpen: boolean;
  /** Seat index of the first player to grab this round, or -1 until one grabs. */
  firstGrabberIdx: number;
  /** Seat index of the player who missed out this round, or -1 until decided. */
  roundLoserIdx: number;
  /** Current round number (1-based). */
  roundNumber: number;
  /** Cards remaining in the feeder's draw pile. */
  drawPileSize: number;
  players: SpoonsPlayer[];
  cpuDifficulty: number;
}
