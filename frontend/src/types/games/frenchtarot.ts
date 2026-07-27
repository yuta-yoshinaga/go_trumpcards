// Type declarations for frenchtarot. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** French Tarot game phase (0=Bid 1=Chien/écart 2=Play 3=TrickEnd 4=RoundEnd 5=GameEnd). */
export type FrenchTarotPhaseValue = 0 | 1 | 2 | 3 | 4 | 5;

/** A French Tarot player's public/own state. Cards are non-empty only for the human. */
export interface FrenchTarotPlayer {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  trickCount: number;
  /** Card-points captured in tricks so far this deal (French Tarot half-point card values). */
  cardPoints: number;
  /** Cumulative match score of this individual player. */
  score: number;
  /** Whether this player is the declarer (contract holder) this deal. */
  isDeclarer: boolean;
}

/** A card played into the current French Tarot trick. */
export interface FrenchTarotTrickCard {
  playerIdx: number;
  card: Card;
}

/** French Tarot game configuration. */
export interface FrenchTarotConfig {
  cpuDifficulty: number;
  /** Number of deals that make up the match; the highest cumulative score wins. */
  targetDeals: number;
}

/** A suggested hint for French Tarot, computed by the backend. */
export interface FrenchTarotHint {
  /** Suggested bid value during the Bid phase, or null/undefined outside it. */
  bid?: number | null;
  cardIndices: number[];
  /** i18n reason suffix identifier. */
  reason: string;
}

/**
 * Full French Tarot (フレンチタロット) game state returned from the API.
 *
 * French Tarot is a 4-player trick-taking game on the 78-card tarot deck (four
 * 14-card suits, 21 numbered trumps, and the Excuse). The human is seat 0. After
 * the auction (Pass / Petite / Garde / Garde Sans / Garde Contre) the highest
 * bidder becomes the declarer, may exchange the 6-card chien (écart) on a
 * Petite/Garde, then all four play out the tricks. The declarer must reach a
 * bouts-based card-point target to win the deal.
 */
export interface FrenchTarotResponse extends BaseGameResponse {
  players: FrenchTarotPlayer[];
  phase: FrenchTarotPhaseValue;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  leadPlayerIdx: number;
  dealerIdx: number;
  /** Seat index of the player currently to bid (Bid phase). */
  bidPlayerIdx: number;
  /** The highest bid so far (0=none/pass, 1=Petite, 2=Garde, 3=Garde Sans, 4=Garde Contre). */
  highestBid: number;
  /** Seat index of the current highest bidder, or -1. */
  highestBidder: number;
  /** Seat index of the declarer, or -1 until decided. */
  declarerIdx: number;
  /** The declared contract (0=None, 1=Petite, 2=Garde, 3=Garde Sans, 4=Garde Contre). */
  contract: number;
  /** Number of cards in the chien (talon). */
  chienCount: number;
  /** The chien cards — non-empty only when revealed to a human declarer during écart. */
  chien: Card[];
  /** Whether the chien has been revealed. */
  chienRevealed: boolean;
  /** Seat index that receives the chien's stashed card points (declarer or -1). */
  stashOwner: number;
  currentTrick: FrenchTarotTrickCard[];
  /** Cumulative match score per player — [p0, p1, p2, p3]. */
  playerScores: number[];
  /** Seat index of the last trick winner, or -1. */
  lastTrickWinner: number;
  /** Deal outcome (0=None, 1=Win/contract made, 2=Loss/contract failed). */
  outcome: number;
  /** Match result from the human's perspective (-1 lose, 0 none, 1 win). */
  result: number;
  /** Indices in the human's hand that are legal to play (non-empty on human Play turn). */
  playableIndices: number[];
  gameEndFlag: boolean;
  /** Winning player seat index, or -1 until the game ends. */
  winnerPlayer: number;
  /** Whether it is currently the human's turn to act (play). */
  isHumanTurn: boolean;
  /** Whether it is currently the human's turn to bid. */
  isHumanBidTurn: boolean;
  /** Whether it is currently the human's turn to discard the écart (6 cards). */
  isHumanDiscard: boolean;
  hint?: FrenchTarotHint | null;
  config: FrenchTarotConfig;
}

// --- Scarto (Piedmontese Tarot) ---
