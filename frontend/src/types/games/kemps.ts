// Type declarations for kemps. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Kemps phase value: 0=Exchange, 1=Declare, 2=RoundEnd, 3=GameEnd. */
export type KempsPhaseValue = 0 | 1 | 2 | 3;

/**
 * A single Kemps player as returned from the API.
 *
 * Four seats split into two teams (even seats = Team A, odd seats = Team B).
 * Only the human (seat 0) has a populated `hand`; CPU hands are an empty array.
 */
export interface KempsPlayer {
  /** Display name ("あなた" / "CPU"). */
  name: string;
  isHuman: boolean;
  /** Team number (0 = Team A, 1 = Team B). */
  team: number;
  /** Number of cards currently held (always 4 during play). */
  handSize: number;
  /** The player's cards — populated only for the human. */
  hand: Card[];
  /** Whether this player currently holds four of a kind (human only). */
  hasFourOfAKind: boolean;
}

/**
 * Full Kemps game state returned from the API.
 *
 * Kemps is a 4-player, 2-team matching game. Each turn a player swaps one hand
 * card for a card on the shared 4-card field, trying to collect four of a kind.
 * When the human's partner secretly signals, the team races to declare "Kemps!"
 * for a point; the opposing team can call "Counter-Kemps!" against a seat to
 * steal it (−1 if wrong). First team to the target score (default 5) wins.
 */
export interface KempsResponse extends BaseGameResponse {
  phase: KempsPhaseValue;
  gameEndFlag: boolean;
  /** Winning team (0 or 1), or -1 until the game ends. */
  winnerTeam: number;
  /** Seat index whose turn it currently is. */
  currentPlayerIdx: number;
  /** Whether it is currently the human's turn to act. */
  isHumanTurn: boolean;
  /** Team scores indexed by team number (Team A = index 0). */
  teamScores: number[];
  /** The shared field of cards available to swap with. */
  field: Card[];
  /** The human's secret signal type (0=Sound, 1=Blink). */
  signalType: number;
  /** Whether the human's partner is currently signaling (human-only cue). */
  partnerSignaling: boolean;
  /** Whether an opponent may be signaling (vague human-only cue). */
  opponentSignaling: boolean;
  /** Seat index that completed four of a kind, or -1. */
  fourHolderIdx: number;
  /** Round result code (0=none, 1=Kemps, 2=Counter, 3=CounterFail, 4=Miss). */
  roundResult: number;
  /** Team that won the most recent round, or -1. */
  roundWinnerTeam: number;
  /** Current round number (1-based). */
  roundNumber: number;
  players: KempsPlayer[];
  cpuDifficulty: number;
  /** Team score required to win (default 5). */
  targetScore: number;
}
