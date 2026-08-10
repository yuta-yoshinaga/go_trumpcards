// Type declarations for macau. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Macau player data with scores and declaration state. */
export interface MacauPlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  roundScore: number;
  cumulativeScore: number;
  hasDeclared: boolean;
}

/** Macau game configuration. */
export interface MacauConfig {
  cpuDifficulty: number;
  pointLimit: number;
}

/** Full Macau game state returned from the API. */
export interface MacauResponse extends BaseGameResponse {
  players: MacauPlayerData[];
  phase: number;
  roundNumber: number;
  currentPlayerIdx: number;
  discardTop: Card | null;
  drawPileCount: number;
  chosenSuit: number;
  penaltyDrawCount: number;
  /**
   * いま出せる手札の位置。
   *
   * マジックカード (2/7/8/J) やチョウズドスートが絡む合法判定を、CUI は
   * HintOutput で全部並べているのに Web は都度クリックで確かめるしかなかった
   * (#4805)。自分の手番でないときは空。
   */
  playableIndices: number[];
  direction: number;
  gameEndFlag: boolean;
  winnerIdx: number;
  config: MacauConfig;
}

// --- Mao (マオ) ---
