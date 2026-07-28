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
}
