// Type declarations for ombre. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Ombre game phase (0=Bid 1=Play 2=TrickEnd 3=RoundEnd 4=GameEnd). */
export type OmbrePhaseValue = 0 | 1 | 2 | 3 | 4;

/** An Ombre player's public/own state. Cards are non-empty only for the human during play. */
export interface OmbrePlayer {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  trickCount: number;
  /** Cumulative match score of this individual player. */
  score: number;
  /** Whether this player is the round's Ombre (won the bid, plays alone). */
  isOmbre: boolean;
}

/** A card played into the current Ombre trick. */
export interface OmbreTrickCard {
  playerIdx: number;
  card: Card;
}

/** Ombre game configuration. */
export interface OmbreConfig {
  cpuDifficulty: number;
  /** Number of deals that make up the match; the highest cumulative score wins. */
  targetRounds: number;
}

/** A suggested hint for Ombre, computed by the backend. */
export interface OmbreHint {
  cardIndices: number[];
  /** i18n reason suffix identifier. */
  reason: string;
}

/**
 * Full Ombre (Hombre) game state returned from the API.
 *
 * Ombre is a 3-player soloist-vs-coalition trick-taker on a 40-card Spanish
 * deck (no 8/9/10). A Bid phase (pass / entrar / solo) plus a chosen trump suit
 * decides the Ombre, who then plays alone against the coalition of the other
 * two. The trump group ranks Spadille (♠A) > Manille (7 of trump) > Basto (♣A)
 * > Punto (Ace of trump, red only) > K > Q > J > 6..2 of trump.
 */
export interface OmbreResponse extends BaseGameResponse {
  players: OmbrePlayer[];
  phase: OmbrePhaseValue;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  /** Seat index of the player whose turn it is to bid. */
  currentBidderIdx: number;
  leadPlayerIdx: number;
  dealerIdx: number;
  /** Seat index of the forehand (first to bid / lead). */
  forehandIdx: number;
  /** Seat index of the round's Ombre (bid winner), or -1 until decided. */
  ombreIdx: number;
  /** The winning bid (0=pass/none, 1=entrar, 2=solo). */
  winningBid: number;
  /** The trump suit (1=♠ 2=♣ 3=♥ 4=♦), or -1 until chosen. */
  trumpSuit: number;
  currentTrick: OmbreTrickCard[];
  /** Cumulative match scores per player — [p0, p1, p2]. */
  playerScores: number[];
  /** Seat index of the last trick winner, or -1. */
  lastTrickWinner: number;
  /** Deal outcome (0=None, 1=Sacar, 2=Puesta, 3=Codille). */
  outcome: number;
  /** Match result from the human's perspective (-1 lose, 0 none, 1 win). */
  result: number;
  /** Indices in the human's hand that are legal to play (non-empty on human Play turn). */
  playableIndices: number[];
  gameEndFlag: boolean;
  /** Winning player seat index, or -1 until the game ends. */
  winnerPlayer: number;
  /** Whether it is currently the human's turn to play a card. */
  isHumanTurn: boolean;
  /** Whether it is currently the human's turn to bid. */
  isHumanBidTurn: boolean;
  hint?: OmbreHint | null;
  config: OmbreConfig;
}

// --- Ulti (Ultimo) ---
