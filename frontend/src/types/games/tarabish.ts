// Type declarations for tarabish. Split-file layout introduced by issue
// #4366; card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** A card played into the current Tarabish trick. */
export interface TarabishTrickCard {
  playerIdx: number;
  card: Card;
}

/** One seat at a Tarabish table. */
export interface TarabishPlayer {
  id: number;
  isHuman: boolean;
  /** `0` or `1`. Seats 0+2 are one partnership, 1+3 the other. */
  team: number;
  /** Hand size. The only hand information exposed for the CPU. */
  cardCount: number;
  /** Populated for the human player only; empty for the CPU. */
  cards: Card[];
  /** Meld total computed from the deal (runs plus bella). */
  meldPoints: number;
  /** Longest run declared, or 0. */
  runLen: number;
  /** Holds the trump K and Q. */
  hasBella: boolean;
  trickCount: number;
}

/**
 * A suggestion. During bidding it advises whether to take trump and carries no
 * `cardIndex`; during play it names a card.
 */
export interface TarabishHint {
  cardIndex?: number;
  /**
   * `tarabishTakeTrump` / `tarabishPassTrump` while bidding;
   * `tarabishWinTrick` or `tarabishFeedPartner` (your partner is winning —
   * throw points on it) during play.
   */
  reason: string;
}

/** Target-score setting. */
export interface TarabishConfig {
  /** Points needed to win (100..1000, default 500). */
  target: number;
}

/** Full Tarabish game state returned from the API. */
export interface TarabishResponse extends BaseGameResponse {
  players: TarabishPlayer[];
  /** `0` = Bid, `1` = Play, `2` = RoundEnd, `3` = GameEnd. */
  phase: number;
  roundNumber: number;
  trickNumber: number;
  trumpSuit: number;
  /** The card turned to propose trump. It joins the dealer's hand once trump is settled. */
  upCard?: Card;
  /** Seat that took trump, or `-1` before it is settled. */
  trumpTakerIdx: number;
  currentPlayerIdx: number;
  leadPlayerIdx: number;
  dealerIdx: number;
  /** Running total per team, index 0 and 1. */
  scores: number[];
  /** Points taken this round per team, index 0 and 1. */
  roundPoints: number[];
  currentTrick: TarabishTrickCard[];
  /** Hand indices you may legally play. Following suit is compulsory. */
  validPlays: number[];
  gameEndFlag: boolean;
  /** `-1` until decided, and `-1` on a tie. */
  winnerTeam: number;
  hint?: TarabishHint;
  config: TarabishConfig;
}
