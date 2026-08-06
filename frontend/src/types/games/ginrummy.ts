// Type declarations for ginrummy. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Gin Rummy player data with scores. */
export interface GinRummyPlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  roundScore: number;
  cumulativeScore: number;
}

/** A meld (set or run) in Gin Rummy. */
export interface GinRummyMeld {
  cards: Card[];
}

/** Gin Rummy game configuration. */
export interface GinRummyConfig {
  cpuDifficulty: number;
  pointLimit: number;
}

// --- Indian Rummy (インドラミー) ---

/** Full Gin Rummy game state returned from the API. */
export interface GinRummyResponse extends BaseGameResponse {
  players: GinRummyPlayerData[];
  phase: number;
  roundNumber: number;
  currentPlayerIdx: number;
  discardTop: Card | null;
  drawPileCount: number;
  gameEndFlag: boolean;
  winnerIdx: number;
  knockerIdx: number;
  knockerMelds: GinRummyMeld[];
  /**
   * 手札のカードごとに、それを足せるノッカーのメルド番号。
   *
   * レイオフフェーズの主題そのものなのに、押してサーバーの応答で初めて成否が
   * 分かる状態だった (#4823)。
   */
  layoffTargets: number[][];
  knockerDeadwood: Card[];
  isGin: boolean;
  config: GinRummyConfig;
}

// --- Conquian (コンキャン) ---
