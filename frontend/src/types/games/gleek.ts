// Type declarations for gleek. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/**
 * Gleek phase values (sync: internal/domain/Gleek.go).
 *
 * 0=Bid, 1=Discard, 2=Play, 3=TrickEnd, 4=RoundEnd, 5=GameEnd. The Discard
 * phase belongs to the buyer alone — every other seat waits through it.
 */
export type GleekPhaseValue = 0 | 1 | 2 | 3 | 4 | 5;

/** A Gleek player's public/own state. Cards are non-empty only for the human. */
export interface GleekPlayer {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  trickCount: number;
  /** Cumulative match score of this individual player. */
  score: number;
  /** Whether this seat bought the stock this deal. */
  isBuyer: boolean;
  /** The amount this seat has bid; 0 if it never bid. */
  bid: number;
  /** Whether this seat has dropped out of the auction. */
  passed: boolean;
  /** Points taken in tricks this deal (3 per trick plus trump honours). */
  trickPoints: number;
  /** This seat's best single-suit total (A=11, courts=10, numerals at face value). */
  ruff: number;
  /** The suit that total came from (1=♠ 2=♣ 3=♥ 4=♦); -1 before the ruff is scored. */
  ruffSuit: number;
}

/** A card played into the current Gleek trick. */
export interface GleekTrickCard {
  playerIdx: number;
  card: Card;
}

/** One declared gleek (three of a rank) or mournival (four). */
export interface GleekMeld {
  playerIdx: number;
  /** 1=ace, 13=king, 12=queen, 11=jack. */
  rank: number;
  /** 3 for a gleek, 4 for a mournival. */
  count: number;
  /** Points taken from each opponent. */
  value: number;
}

/** Gleek game configuration. */
export interface GleekConfig {
  cpuDifficulty: number;
  /** Number of deals that make up the match; the highest cumulative score wins. */
  targetRounds: number;
}

/** A suggested hint for Gleek, computed by the backend. */
export interface GleekHint {
  cardIndices: number[];
  /** i18n reason suffix identifier. */
  reason: string;
}

/**
 * Full Gleek game state returned from the API.
 *
 * Gleek is a 16th–17th century English three-hander on a 44-card pack that runs
 * four scoring stages inside one deal: an auction for the stock, the ruff, the
 * gleek/mournival melds, and twelve tricks scored 3 apiece plus trump honours.
 */
export interface GleekResponse extends BaseGameResponse {
  players: GleekPlayer[];
  phase: GleekPhaseValue;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  /** Seat index of the player whose turn it is to bid. */
  currentBidderIdx: number;
  leadPlayerIdx: number;
  dealerIdx: number;
  /** Seat left of the dealer; opens the auction and leads the first trick. */
  elderIdx: number;
  /** Seat that bought the stock; -1 until the auction closes. */
  buyerIdx: number;
  winningBid: number;
  /** Highest amount standing in the auction. */
  highestBid: number;
  /** The only amount this seat may bid; 0 once the ceiling is reached. */
  nextBidAmount: number;
  /** Trump suit, fixed by the turn-up before the auction starts (1=♠ 2=♣ 3=♥ 4=♦). */
  trumpSuit: number;
  /** The card turned up to fix trump. It never enters a trick. */
  turnUp?: Card | null;
  currentTrick: GleekTrickCard[];
  /** Cumulative match scores per player — [p0, p1, p2]. */
  playerScores: number[];
  /** How many cards the buyer must throw (seven). */
  discardCount: number;
  /** Seat that won the ruff; -1 until it is scored. */
  ruffWinnerIdx: number;
  melds: GleekMeld[];
  /** Points this deal actually held (72–81; the turn-up and discards never reach the table). */
  dealPoints: number;
  /** dealPoints divided by three — what each seat is settled against. */
  par: number;
  /** Seat index of the last trick winner, or -1. */
  lastTrickWinner: number;
  /** Match result from the human's perspective (-1 lose, 0 none, 1 win). */
  result: number;
  /** Indices in the human's hand that are legal to play (non-empty on human Play turn). */
  playableIndices: number[];
  gameEndFlag: boolean;
  /** Winning player seat index, or -1 until the game ends. */
  winnerPlayer: number;
  isHumanTurn: boolean;
  isHumanBidTurn: boolean;
  /** Whether the human bought the stock and still owes seven discards. */
  isHumanDiscardTurn: boolean;
  hint?: GleekHint | null;
  config: GleekConfig;
}
