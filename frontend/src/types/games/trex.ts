// Type declarations for trex. Split-file layout introduced by issue #4366;
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** One Trex seat. */
export interface TrexPlayer {
  id: number;
  isHuman: boolean;
  /** Hand size. Always sent, including while {@link TrexPlayer.hidden} is true. */
  cardCount: number;
  /** Empty while {@link TrexPlayer.hidden} is true. */
  cards: Card[];
  /**
   * Running total. Public for every seat — who is down and by how much is what
   * a king weighs when choosing. Can be positive, because the dominoes pay out.
   */
  score: number;
  dealScore: number;
  tricksWon: number;
  hidden: boolean;
}

/** One card played into the current trick. */
export interface TrexTrickCard {
  playerIdx: number;
  card: Card;
}

/** One suit's dominoes run. */
export interface TrexRun {
  suit: number;
  started: boolean;
  /** Range on the table, ace high (14). The jack that starts a run is 11. */
  low: number;
  high: number;
}

/** Suggested move for the human seat. */
export interface TrexHintPayload {
  cardIndex?: number;
  contract?: number;
  pass: boolean;
  /** Reason identifier, e.g. `trex.hint.choose`. */
  reason: string;
}

/** Full Trex game state returned from the API. */
export interface TrexResponse extends BaseGameResponse {
  players: TrexPlayer[];
  /** 0 = Choose, 1 = Play, 2 = DealEnd, 3 = GameEnd. */
  phase: number;
  currentPlayerIdx: number;
  /** Seat choosing the contract for this deal. */
  kingIdx: number;
  /** 0 = KingOfHearts, 1 = Diamonds, 2 = Queens, 3 = Tricks, 4 = Dominoes, 5 = none. */
  contract: number;
  /**
   * Contracts this king has not played yet. A contract is played once per
   * kingdom, so the page never has to track what is spent.
   */
  availableContracts: number[];
  /** The dominoes contract, which builds runs instead of tricks. */
  isTrix: boolean;
  dealNo: number;
  /** Twenty — four kingdoms of five contracts. */
  totalDeals: number;
  /** Empty during the dominoes contract. */
  trick: TrexTrickCard[];
  trickNo: number;
  /** The four dominoes runs, one per suit. */
  runs: TrexRun[];
  /** Seats in the order they went out during the dominoes. */
  finishOrder: number[];
  /**
   * Hand indices that may be played. Each contract follows a different rule, so
   * this is computed once server-side.
   */
  validIndices: number[];
  /**
   * True only in the dominoes with no legal play. Trick contracts have no pass
   * at all.
   */
  canPass: boolean;
  gameEndFlag: boolean;
  /** Highest final score; -1 before the twentieth deal. */
  winnerIdx: number;
  hint?: TrexHintPayload;
}
