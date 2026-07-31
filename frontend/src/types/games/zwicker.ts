// Type declarations for zwicker. Split-file layout introduced by issue #4366;
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/**
 * A card together with every matching value it can take.
 *
 * The server sends the values rather than letting the client derive them:
 * **aces and court cards each have two** (A 1/11, J 2/12, Q 3/13, K 4/14) and
 * the three jokers are fixed at 15/20/25, so a client-side copy of that table
 * would drift from the server's capture check.
 */
export interface ZwickerCard extends Card {
  values: number[];
}

/** One Zwicker seat. */
export interface ZwickerPlayer {
  id: number;
  isHuman: boolean;
  /** 0 or 1. Seats across the table are partners. */
  team: number;
  /** Hand size. Always sent, including while {@link ZwickerPlayer.hidden} is true. */
  cardCount: number;
  /** Empty while {@link ZwickerPlayer.hidden} is true. */
  cards: ZwickerCard[];
  /**
   * Cards taken this deal. Public for every seat: the majority of cards is
   * worth three, so this count is what the decision is made against.
   */
  capturedCount: number;
  /** Times this seat cleared the table. One point each. */
  zwicks: number;
  hidden: boolean;
}

/** One build on the table. */
export interface ZwickerBuild {
  owner: number;
  /** Declared value. A build is captured only at exactly this value. */
  value: number;
  cards: Card[];
}

/** Breakdown of the deal just scored. */
export interface ZwickerRoundScore {
  cardPoints: number[];
  cards: number[];
  /** Team with more cards; -1 when the counts were level and nobody takes the three. */
  majorityTeam: number;
  zwicks: number[];
  total: number[];
}

/** Suggested move for the human seat. */
export interface ZwickerHintPayload {
  /** True to capture, false to trail. */
  take: boolean;
  cardIndex?: number;
  value?: number;
  tableIndices?: number[];
  /** Reason identifier, e.g. `zwicker.hint.take`. */
  reason: string;
}

/** Full Zwicker game state returned from the API. */
export interface ZwickerResponse extends BaseGameResponse {
  players: ZwickerPlayer[];
  /** 0 = Play, 1 = RoundEnd, 2 = GameEnd. */
  phase: number;
  currentPlayerIdx: number;
  stockCount: number;
  tableCards: ZwickerCard[];
  builds: ZwickerBuild[];
  /** Running totals, indexed by team. */
  teamScores: number[];
  /** Points that win the game. */
  targetScore: number;
  /** Breakdown of the deal just scored; absent until one is. */
  lastRound?: ZwickerRoundScore;
  gameEndFlag: boolean;
  /** Winning team; -1 while the game is live. */
  winnerTeam: number;
  hint?: ZwickerHintPayload;
}
