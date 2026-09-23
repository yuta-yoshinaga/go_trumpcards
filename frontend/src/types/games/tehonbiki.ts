import type { BaseGameResponse } from '../common';

/** Number wager kinds and their adopted house-table profit ratios. */
export const TEHONBIKI_PAYOUTS = { single: [9, 2], double: [9, 5], triple: [9, 10], half: [9, 10] } as const;
/** Response payload for the Tehonbiki endpoint. */
export interface TehonbikiResponse extends BaseGameResponse {
  phase: number;
  parentCard?: number;
  numbers: number[];
  betType: keyof typeof TEHONBIKI_PAYOUTS | '';
  bet: number;
  result: number;
  payout: number;
  chips: number;
  roundNumber: number;
  remainingCards: number;
  gameEndFlag: boolean;
  payoutNum: number;
  payoutDen: number;
  config?: { initialChips: number; defaultBet: number };
}
