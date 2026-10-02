// Type declarations for pineapple. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { Card } from '../common';
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
  discardPreviews?: PineappleDiscardPreview[];
}

/** Server-evaluated discard option for Irish Poker or Crazy Pineapple. */
export interface PineappleDiscardPreview {
  /** Original four-card hand indices (Irish Poker). */
  discardIdx0?: number;
  discardIdx1?: number;
  /** Original hand index (Crazy Pineapple). */
  cardIdx?: number;
  handRank: number;
  recommended?: boolean;
  strengthCards?: Card[];
  /** Relative strength among all Irish Poker pairs; larger is stronger. */
  strengthOrder?: number;
}
