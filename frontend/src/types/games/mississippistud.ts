// Type declarations for mississippistud. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card, MaskedCard } from '../common';

/** Mississippi Stud API response. */
export interface MississippiStudResponse extends BaseGameResponse {
  /** Player's two hole cards (revealed once the round starts). */
  playerHand: Card[];
  /** Community cards: masked as `MaskedCard` until the matching street is revealed. */
  communityCards: (Card | MaskedCard)[];
  /** Per-card reveal state for `communityCards` (length 3). */
  communityRevealed: boolean[];
  phase: number;
  chips: number;
  anteAmount: number;
  /** 3rd / 4th / 5th street bet multipliers (0=未ベット, 1/2/3=倍率). Length 3. */
  streetMultipliers: number[];
  folded: boolean;
  totalBet: number;
  result: number;
  handRank: number;
  /** Applied payout multiplier (-1=push, 0=loss, positive=win). */
  payoutMultiplier: number;
  antePayout: number;
  streetPayouts: number[];
  totalPayout: number;
}

// --- Pai Gow Poker (パイゴウポーカー) ---
