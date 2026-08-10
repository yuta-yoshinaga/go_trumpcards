// Type declarations for barbu. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Per-game configuration for Barbu. */
export interface BarbuConfig {
  cpuDifficulty: number;
}

/** A single Barbu player's view. Cards are populated only for the human. */
export interface BarbuPlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  trickCount: number;
  dominoRank: number;
  totalScore: number;
}

/** A single card played into the current/last trick. */
export interface BarbuTrickCard {
  playerIdx: number;
  card: Card;
}

/** One deal's scoring breakdown. */
export interface BarbuDealDetail {
  contract: number;
  trumpSuit: number;
  dealerIdx: number;
  gained: Record<number, number>;
}

/** API response shape for a Barbu game. */
export interface BarbuResponse extends BaseGameResponse {
  players: BarbuPlayerData[];
  phase: string;
  dealNumber: number;
  totalDeals: number;
  dealerIdx: number;
  currentTurn: number;
  currentContract: number;
  trumpSuit: number;
  trickNumber: number;
  currentTrick: BarbuTrickCard[];
  lastTrick: BarbuTrickCard[];
  lastTrickWinner: number;
  tablePlaced: number[];
  dominoPlayable: number[];
  /**
   * トリック契約でいま出せる手札の位置（フォロー義務を反映）。
   *
   * ドミノ契約以外では可視化が無く、リード色を持っていても全カードが押せて、
   * サーバーに弾かれて初めて分かる状態だった (#4804)。人間の手番でないときは空。
   */
  playableIndices: number[];
  usedContracts: boolean[];
  gameEndFlag: boolean;
  config: BarbuConfig;
  roundWinners: number[];
  lastDealDetail: BarbuDealDetail | null;
  dealHistory: BarbuDealDetail[];
}

// --- Spite and Malice (スパイト・アンド・マリス) ---
