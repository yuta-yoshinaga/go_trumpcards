// Type declarations for chinesepoker. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Chinese Poker game state response. */
export interface ChinesePokerResponse extends BaseGameResponse {
  playerCards: Card[];
  dealerCards: Card[];
  playerFront: Card[];
  playerMiddle: Card[];
  playerBack: Card[];
  dealerFront: Card[];
  dealerMiddle: Card[];
  dealerBack: Card[];
  phase: number;
  chips: number;
  bet: number;
  result: number;
  frontResult: number;
  middleResult: number;
  backResult: number;
  payout: number;
  playerFrontRank: number;
  playerMiddleRank: number;
  playerBackRank: number;
  dealerFrontRank: number;
  dealerMiddleRank: number;
  dealerBackRank: number;
  playerRoyalty: number;
  dealerRoyalty: number;
  scoop: boolean;
  /**
   * The split the server recommends, as indices into `playerCards`.
   *
   * Present only during SET_HANDS with a full hand. The CUI has printed this
   * since #4717; the web page used to derive its own ranking instead, so the
   * two surfaces could recommend different splits for the same hand (#5615).
   */
  suggestedArrangement?: {
    front: number[];
    middle: number[];
    back: number[];
    /** Whether that split fouls (front > middle, or middle > back). */
    foul: boolean;
  };
}
