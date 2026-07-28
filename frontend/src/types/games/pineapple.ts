// Type declarations for pineapple. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { HoldemResponse } from './holdem';

/** Pineapple Poker response extending Hold'em with discard phase fields. */
export interface PineappleResponse extends HoldemResponse {
  isDiscardPhase: boolean;
  discardDone: boolean[];
  initialDealCount: number;
}
