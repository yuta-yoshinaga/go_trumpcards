// Type declarations for casinowar. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Casino War API response. */
export interface CasinoWarResponse extends BaseGameResponse {
  /** Player's initial card. */
  playerCard?: Card;
  /** Dealer's initial card. */
  dealerCard?: Card;
  /** Player's war card (only set after going to war). */
  playerWarCard?: Card;
  /** Dealer's war card (only set after going to war). */
  dealerWarCard?: Card;
  /** Burn cards face-down between initial and war (length 0 or 3). */
  burnCards: Card[];
  phase: number;
  chips: number;
  ante: number;
  /** Additional bet placed when going to war (equal to ante). */
  warBet: number;
  result: number;
  totalPayout: number;
}

// --- Oicho-Kabu (おいちょかぶ) ---
