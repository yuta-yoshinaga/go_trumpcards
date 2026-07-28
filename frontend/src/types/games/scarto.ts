// Type declarations for scarto. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Scarto game phase (0=Scarto/discard 1=Play 2=TrickEnd 3=RoundEnd 4=GameEnd). */
export type ScartoPhaseValue = 0 | 1 | 2 | 3 | 4;

/** A Scarto player's public/own state. Cards are non-empty only for the human. */
export interface ScartoPlayer {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  trickCount: number;
  /** Card-points captured in tricks so far this deal (Italian tarocchi card values). */
  cardPoints: number;
  /** Cumulative match score of this individual player. */
  score: number;
  /** Whether this player is the dealer this deal (the dealer performs the scarto). */
  isDealer: boolean;
}

/** A card played into the current Scarto trick. */
export interface ScartoTrickCard {
  playerIdx: number;
  card: Card;
}

/** Scarto game configuration. */
export interface ScartoConfig {
  cpuDifficulty: number;
  /** Number of deals that make up the match; the highest cumulative score wins. */
  targetDeals: number;
}

/** A suggested hint for Scarto, computed by the backend. */
export interface ScartoHint {
  cardIndices: number[];
  /** i18n reason suffix identifier. */
  reason: string;
}

/**
 * Full Scarto (スカルト) game state returned from the API.
 *
 * Scarto is a simple 3-player Italian tarocchi trick-taker on the 78-card tarot
 * deck (four 14-card suits, 21 numbered trumps, and the Excuse). The human is
 * seat 0. There is no bidding, no chien, and no partnership: the dealer buries
 * three low pip cards (the scarto), then the three players play trump-priority
 * tricks. Each deal is scored as a zero-sum settlement against the average of
 * the three players' captured card-points; the highest cumulative score after
 * the set number of deals wins.
 */
export interface ScartoResponse extends BaseGameResponse {
  players: ScartoPlayer[];
  phase: ScartoPhaseValue;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  leadPlayerIdx: number;
  /** Seat index of the dealer, who performs the scarto (discard). */
  dealerIdx: number;
  /** Number of cards the dealer has already buried this deal (0 until the scarto is done, then 3). */
  scartoCount: number;
  currentTrick: ScartoTrickCard[];
  /** Cumulative match score per player — [p0, p1, p2]. */
  playerScores: number[];
  /** Signed settlement of the most recent deal per player — [p0, p1, p2]. */
  dealScores: number[];
  /** Seat index of the last trick winner, or -1. */
  lastTrickWinner: number;
  /** Deal outcome from the human's perspective (0=None/average, 1=above average, 2=below average). */
  outcome: number;
  /** Match result from the human's perspective (-1 lose, 0 none, 1 win). */
  result: number;
  /** Indices in the human's hand that are legal to play (non-empty on human Play turn). */
  playableIndices: number[];
  gameEndFlag: boolean;
  /** Winning player seat index, or -1 for a draw / undecided. */
  winnerPlayer: number;
  /** Whether it is currently the human's turn to act (play). */
  isHumanTurn: boolean;
  /** Whether it is currently the human's turn to perform the scarto (they are the dealer). */
  isHumanScarto: boolean;
  hint?: ScartoHint | null;
  config: ScartoConfig;
}

// --- Königrufen (Tarock) ---
