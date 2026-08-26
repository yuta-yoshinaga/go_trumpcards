// Type declarations for piedmontesetarot. Follows the split-out convention of
// card.ts (issue #4366); card.ts re-exports this file.

import type { BaseGameResponse, Card } from '../common';

/** Piedmontese Tarot phase (0=Scarto 1=Play 2=TrickEnd 3=RoundEnd 4=GameEnd). */
export type PiedmonteseTarotPhaseValue = 0 | 1 | 2 | 3 | 4;

/** A seat's public/own state. Cards are non-empty only for the human. */
export interface PiedmonteseTarotPlayer {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  trickCount: number;
  /**
   * Card points captured so far this deal, in thirds.
   *
   * The classic Italian count works in groups of three cards, each group worth
   * its values minus two, so a seat's share is not always a whole number. The
   * server keeps it in thirds and sends the readable form in `cardPoints`.
   */
  cardThirds: number;
  /** The same figure written for the screen, e.g. "26 1/3". */
  cardPoints: string;
  /** Cumulative match score of this seat. */
  score: number;
  /** Whether this seat deals — and therefore performs the scarto — this deal. */
  isDealer: boolean;
}

/** A card played into the current trick. */
export interface PiedmonteseTarotTrickCard {
  playerIdx: number;
  card: Card;
}

/** Piedmontese Tarot game configuration. */
export interface PiedmonteseTarotConfig {
  /** Table size: 3 or 4. The deal and the talon change with it. */
  seats: number;
  cpuDifficulty: number;
  /** Number of deals that make up the match; the highest cumulative score wins. */
  targetDeals: number;
}

/** A suggested hint, computed by the backend. */
export interface PiedmonteseTarotHint {
  cardIndices: number[];
  /** i18n reason identifier. */
  reason: string;
}

/**
 * Full Tarocco Piemontese game state returned from the API.
 *
 * The Piedmontese four-player tarot on the 78-card French-suited deck (four
 * 14-card suits, 21 trumps and the Matto). The human is seat 0. The dealer
 * picks up the talon and buries the same number of cards (the scarto), then
 * the table plays trump-priority tricks. Each deal settles zero-sum against
 * the table's captured card points; the highest cumulative score after the set
 * number of deals wins.
 */
export interface PiedmonteseTarotResponse extends BaseGameResponse {
  players: PiedmonteseTarotPlayer[];
  phase: PiedmonteseTarotPhaseValue;
  roundNumber: number;
  trickNumber: number;
  /** Tricks in a deal — equal to the hand size (19 at four seats, 25 at three). */
  trickCount: number;
  currentPlayerIdx: number;
  leadPlayerIdx: number;
  /** Seat index of the dealer, who performs the scarto. */
  dealerIdx: number;
  /** Cards the dealer has already buried this deal (0 until the scarto is done). */
  scartoCount: number;
  /** Cards the dealer must bury: 2 at four seats, 3 at three seats. */
  talonSize: number;
  currentTrick: PiedmonteseTarotTrickCard[];
  /** Cumulative match score per seat. */
  playerScores: number[];
  /** Signed settlement of the most recent deal per seat. */
  dealScores: number[];
  /** Seat index of the last trick winner, or -1. */
  lastTrickWinner: number;
  /** Deal outcome from the human's perspective (0=none, 1=above average, 2=below). */
  outcome: number;
  /** Match result from the human's perspective (0=none/draw, 1=win, 2=lose). */
  result: number;
  /** Indices in the human's hand that are legal to play. */
  playableIndices: number[];
  /**
   * Indices in the dealer's hand that may be buried in the scarto.
   *
   * **Computed by the domain, not derivable from the card faces.** When the
   * dealer holds too few freely-buriable pips, ordinary (non-bout) trumps
   * become legal, and a colour-based rule on the client cannot see that — which
   * is why the dealer could not make up the count from the screen (#6236).
   */
  discardableIndices: number[];
  gameEndFlag: boolean;
  /** Winning seat index, or -1 for a draw / undecided. */
  winnerPlayer: number;
  /** Whether it is the human's turn to play. */
  isHumanTurn: boolean;
  /** Whether it is the human's turn to perform the scarto (they are the dealer). */
  isHumanScarto: boolean;
  hint?: PiedmonteseTarotHint | null;
  config: PiedmonteseTarotConfig;
}
