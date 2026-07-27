// Type declarations for dragontiger. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Dragon Tiger game state response. Bet types: 0=Dragon, 1=Tiger, 2=Tie. */
export interface DragonTigerResponse extends BaseGameResponse {
  /** Card dealt to the Dragon slot. */
  dragonCard?: Card;
  /** Card dealt to the Tiger slot. */
  tigerCard?: Card;
  phase: number;
  chips: number;
  betAmount: number;
  /** 0=Dragon, 1=Tiger, 2=Tie */
  betType: number;
  /** Domain GameResult on the wire: 1=Dragon wins, -1=Tiger wins, 0=Tie */
  result: number;
  payout: number;
  /** Big Road history. 0=Dragon, 1=Tiger, 2=Tie. */
  history: number[];
}
