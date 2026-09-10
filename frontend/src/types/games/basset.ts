import type { BaseGameResponse, Card } from '../common';

/** A wager currently held on one Basset rank. */
export interface BassetBet {
  rank: number;
  amount: number;
  stage: number;
}

/** Server response for the Basset game. */
export interface BassetResponse extends BaseGameResponse {
  phase: number;
  chips: number;
  bet: BassetBet | null;
  bankerCard: Card | null;
  playerCard: Card | null;
  hit: boolean;
  turnsPlayed: number;
  turnsTotal: number;
  remaining: number;
  remainingByRank: number[];
  totalPayout: number;
  gameEndFlag: boolean;
}
