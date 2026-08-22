// Type declarations for threecardrummy. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card, MaskedCard } from '../common';

/** Three Card Rummy API response. Scores are totals — **lower is better**. */
export interface ThreeCardRummyResponse extends BaseGameResponse {
  playerHand: Card[];
  /**
   * **End フェーズより前は全枚数がマスク済み** (`design: ''`, `value: 0`)。
   * 3 枚とも見えていたらディーラーの合計が数えられ、play/fold に判断の余地が
   * 無くなる。読む側は `isMaskedCard` で分岐すること。
   */
  dealerHand: (Card | MaskedCard)[];
  phase: number;
  chips: number;
  anteBet: number;
  lowBonusBet: number;
  playBet: number;
  result: number;
  antePayout: number;
  playPayout: number;
  anteBonusPayout: number;
  lowBonusPayout: number;
  totalPayout: number;
  dealerQualified: boolean;
  playerScore: number;
  dealerScore: number;
}
