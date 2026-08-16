// Type declarations for pineapple. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { HoldemResponse } from './holdem';

/** Pineapple Poker response extending Hold'em with discard phase fields. */
export interface PineappleResponse extends HoldemResponse {
  isDiscardPhase: boolean;
  discardDone: boolean[];
  initialDealCount: number;
  /**
   * i18n key of the human's best hand so far (`"straightFlush"` etc.), or empty
   * before five cards exist, once the hand reaches showdown, or when the human
   * has folded. Decided by the domain's `PeekBestHand` and sent from the server
   * -- Omaha re-derives the same thing in TypeScript, which is the duplication
   * #5601 removed elsewhere (#5488).
   */
  liveBestHand: string;
}
