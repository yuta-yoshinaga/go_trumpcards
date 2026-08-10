// Type declarations for klaberjass. Split-file layout introduced by issue #4366;
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** A declared run of three or more cards in one suit. */
export interface KlaberjassSequence {
  suit: number;
  /** Top card of the run, with the ace as 14. */
  topValue: number;
  length: number;
  /** 20 for three cards, 50 for four or more — five or longer is still 50. */
  points: number;
}

/** One Klaberjass seat. */
export interface KlaberjassPlayer {
  id: number;
  isHuman: boolean;
  cardCount: number;
  /** Your own hand only; the opponent's stays empty until the settlement. */
  cards: Card[];
  /**
   * Empty until the hand is settled — seeing the opponent's runs during play
   * would give away the contest they are about to lose or win.
   */
  sequences: KlaberjassSequence[];
  /** Points taken this deal, melds and bela included. */
  handPoints: number;
  score: number;
  isMaker: boolean;
  isDealer: boolean;
  isCurrentTurn: boolean;
}

/** Full Klaberjass game state returned from the API. */
export interface KlaberjassResponse extends BaseGameResponse {
  players: KlaberjassPlayer[];
  /** 0 = BidTurnUp, 1 = BidFree, 2 = Schmeiss, 3 = Play, 4 = HandEnd, 5 = GameEnd. */
  phase: number;
  dealNumber: number;
  currentPlayerIdx: number;
  bidPlayerIdx: number;
  dealerIdx: number;
  /** 1 = Spade, 2 = Clover, 3 = Heart, 4 = Diamond; 0 while undecided. */
  trumpSuit: number;
  /**
   * The card turned up to propose trump. It is **not dealt to anyone** — it
   * only enters play if it is swapped for the trump seven (the dix).
   */
  turnUpCard: Card | null;
  /** Seat that fixed trump; -1 while undecided. */
  makerIdx: number;
  trick: Card[];
  trickLeaderIdx: number;
  trickNumber: number;
  /**
   * Hand indices the human may legally play. Sent by the server because
   * following suit, trumping when void and **overtrumping a trump lead** are
   * all compulsory — re-deriving that on the client always drifts.
   */
  validPlays: number[];
  /** Seat that won the sequence contest; -1 when nobody scored. */
  sequenceWinner: number;
  /**
   * Seat that took the last trick and its 10-point bonus, or -1 before the
   * hand ends. Without it the settlement panel cannot explain why the hand
   * points do not add up from the declarations alone (#4937).
   */
  lastTrickWinner: number;
  /** Points the last trick is worth, sent by the server so the two sides cannot drift. */
  lastTrickBonus: number;
  /** Seat holding trump K+Q; -1 when neither does. */
  belaHolder: number;
  belaScored: boolean;
  /** Whether the trump seven was exchanged for the turn-up. */
  dixUsed: boolean;
  /**
   * Whether the maker failed. The maker must score **strictly more** than the
   * opponent — a tie is bete, and the opponent takes both totals.
   */
  bete: boolean;
  /** Seat offering to throw the deal in; -1 when none. */
  schmeissBy: number;
  targetScore: number;
  gameEndFlag: boolean;
  winnerIdx: number;
}
