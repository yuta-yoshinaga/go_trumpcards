// Type declarations for goofspiel. Split-file layout introduced by issue #4366;
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** One seat at a Goofspiel table. */
export interface GoofspielPlayer {
  id: number;
  isHuman: boolean;
  /** Bid cards not yet spent. */
  cardCount: number;
  /**
   * The remaining bid cards. **Populated for every seat, CPUs included.**
   *
   * Spent cards go face up, so an opponent's remainder is simply their suit
   * minus what they have played — hiding it would only add busywork.
   */
  cards: Card[];
  score: number;
  /** Whether this seat has committed a bid this round. */
  hasBid: boolean;
  /** The bid, present only after the reveal. */
  revealedBid?: Card;
}

/** A suggestion. */
export interface GoofspielHint {
  cardIndex?: number;
  /** `goofspielMatch`, `goofspielHighPrize`, `goofspielLowPrize`, or `goofspielCarried`. */
  reason: string;
}

/** Table-size and tie settings. */
export interface GoofspielConfig {
  /** Players at the table (2..3, default 2). */
  playerCnt: number;
  /** `0` = a tied prize is discarded, `1` = it carries into the next round. */
  tieRule: number;
}

/** Full Goofspiel game state returned from the API. */
export interface GoofspielResponse extends BaseGameResponse {
  players: GoofspielPlayer[];
  /** `0` = Bid, `1` = Reveal, `2` = GameEnd. */
  phase: number;
  /** Hand indices you may bid. Empty once you have committed this round. */
  validPlays: number[];
  /** The prize on the table, absent once the round is settled. */
  currentPrize?: Card;
  /** Prizes carried over by earlier ties. */
  carriedPrizes: Card[];
  /** Points at stake this round, **including any carry-over**. */
  prizeValue: number;
  prizeRemaining: number;
  /** The seat that won the last round, or `-1` when it was a tie. */
  lastWinnerIdx: number;
  /** Points moved last round. `0` on a tie. */
  lastGained: number;
  roundNumber: number;
  gameEndFlag: boolean;
  /** `-1` until decided. */
  winnerIdx: number;
  hint?: GoofspielHint;
  config: GoofspielConfig;
}
