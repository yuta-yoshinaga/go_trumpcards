// Type declarations for tongits. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Tongits player data with scores. */
export interface TongitsPlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  roundScore: number;
  cumulativeScore: number;
}

/** A meld (set or run) in Tongits. */
export interface TongitsMeld {
  cards: Card[];
}

/** Tongits game configuration. */
export interface TongitsConfig {
  cpuDifficulty: number;
  pointLimit: number;
}

/** Full Tongits game state returned from the API. */
export interface TongitsResponse extends BaseGameResponse {
  players: TongitsPlayerData[];
  phase: number;
  roundNumber: number;
  currentPlayerIdx: number;
  discardTop: Card | null;
  drawPileCount: number;
  gameEndFlag: boolean;
  winnerIdx: number;
  knockerIdx: number;
  knockerMelds: TongitsMeld[];
  knockerDeadwood: Card[];
  opponentMelds: TongitsMeld[];
  opponentDeadwood: Card[];
  isTongits: boolean;
  /**
   * Opponent hand size at or below which knocking risks an undercut. Sent so
   * the warning fires on the same board state in both UIs.
   */
  undercutRiskMax: number;
  isUndercut: boolean;
  /**
   * 1枚捨てて到達できる最小デッドウッド。人間のディスカードフェーズ以外は -1。
   *
   * **-1 は「まだ聞くべき場面でない」印であって 0 ではない。**0 にすると
   * 「デッドウッド0 = 必ずノック可能」と読めてしまう。
   */
  bestDeadwood: number;
  /** ノックできるデッドウッド上限。`domain.TongitsKnockThreshold` をサーバーが送る。 */
  knockThreshold: number;
  config: TongitsConfig;
}

// --- Thirty-One (サーティワン / Scat) ---
