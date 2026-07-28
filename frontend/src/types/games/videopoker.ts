// Type declarations for videopoker. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Full Video Poker game state returned from the API. */
export interface VideoPokerResponse extends BaseGameResponse {
  hand: Card[];
  phase: number;
  chips: number;
  betAmount: number;
  result: number;
  payout: number;
  handRank: number;
  handName: string;
  /**
   * Stable, locale-independent hand key (e.g. `"wildRoyalFlush"`) matching the
   * `payoutTable.name.*` / `videoPokerPayoutRows` row keys. Empty on a losing
   * hand, and absent on responses that predate this field. Preferred over
   * reverse-looking up the English `handName`.
   */
  handKey?: string;
  heldIndices: boolean[];
  variantName: string;
}

// --- Cribbage (クリベッジ) ---
