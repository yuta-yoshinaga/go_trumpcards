// Type declarations for oichokabu. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/**
 * Oicho-Kabu API response.
 *
 * A kabufuda (40-card, values 1–10) baccarat-style banking game: one human
 * "child" vs a CPU "banker". The banker's hand stays hidden (empty array,
 * rank 0) until the round ends. Rank is the sum of card values mod 10 (9 best,
 * 0 worst); ties push and a win pays 1:1.
 */
export interface OichoKabuResponse extends BaseGameResponse {
  /** The child's (player's) cards. */
  playerHand: Card[];
  /** The banker's cards — empty until the round ends. */
  bankerHand: Card[];
  /** The child's rank (sum mod 10), 0–9. */
  playerRank: number;
  /** The banker's rank (sum mod 10) — 0 until the round ends. */
  bankerRank: number;
  phase: number;
  chips: number;
  /** Chips wagered this round. */
  bet: number;
  result: number;
  totalPayout: number;
}
