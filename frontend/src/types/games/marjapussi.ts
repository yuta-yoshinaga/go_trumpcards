// Type declarations for marjapussi. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Marjapussi game phase (0=Play 1=TrickEnd 2=RoundEnd 3=GameEnd). */
export type MarjapussiPhaseValue = 0 | 1 | 2 | 3;

/** A Marjapussi player's state. */
export interface MarjapussiPlayer {
  id: number;
  teamId: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  trickCount: number;
  score: number;
}

/** A card played into the current Marjapussi trick. */
export interface MarjapussiTrickCard {
  playerIdx: number;
  card: Card;
}

/** Marjapussi game configuration. */
export interface MarjapussiConfig {
  cpuDifficulty: number;
  targetPoints: number;
}

/** A suggested hint for Marjapussi, computed by the backend. */
export interface MarjapussiHint {
  cardIndices: number[];
  /** i18n reason suffix identifier. */
  reason: string;
}

/**
 * Full Marjapussi game state returned from the API.
 *
 * Marjapussi is a Finnish 4-player 2-vs-2 partnership trick-taker played with a 36-card deck (6-A).
 * Seats 0 and 2 form Team 0; seats 1 and 3 form Team 1.
 * There is no card exchange.
 * Leading a King or Queen while holding both in hand declares a marriage and sets the trump suit.
 * The winner of the 8th (final) trick captures the 4-card "pussi" (berry bag).
 */
export interface MarjapussiResponse extends BaseGameResponse {
  players: MarjapussiPlayer[];
  phase: MarjapussiPhaseValue;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  leadPlayerIdx: number;
  dealerIdx: number;
  /** Trump suit (0=unset, 1=♠ 2=♣ 3=♥ 4=♦). 0 until a marriage is declared. */
  trumpSuit: number;
  currentTrick: MarjapussiTrickCard[];
  /** Cumulative team match scores — [Team0, Team1]. */
  teamScores: number[];
  /** Cumulative match scores per player — [p0, p1, p2, p3]. */
  playerScores: number[];
  /** Card points captured per team this round — [Team0, Team1]. */
  roundCardPoints: number[];
  /** Marriage points scored per team this round — [Team0, Team1]. */
  roundMarriage: number[];
  /** Number of cards in the pussi (berry bag). */
  pussiCount: number;
  /** Cards in the pussi, revealed at RoundEnd / GameEnd. */
  pussi?: Card[];
  /** Team index that won the pussi (0 or 1), or -1 if not yet determined. */
  pussiWinnerTeam: number;
  /** Seat index of the last trick winner, or -1. */
  lastTrickWinner: number;
  /** Indices in the human's hand that are legal to play (non-empty on human Play turn). */
  playableIndices: number[];
  gameEndFlag: boolean;
  /** Winning player seat index, or -1 until the game ends. */
  winnerPlayer: number;
  /** Winning team index (0 or 1), or -1 until the game ends. */
  winnerTeam: number;
  /** Whether it is currently the human's turn to act. */
  isHumanTurn: boolean;
  hint?: MarjapussiHint | null;
  config: MarjapussiConfig;
}

// --- Calabresella (Terziglio) ---
