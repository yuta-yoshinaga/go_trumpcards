// Type declarations for crazyeights. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Crazy Eights player data with scores. */
export interface CrazyEightsPlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  roundScore: number;
  cumulativeScore: number;
}

/** Crazy Eights game configuration. */
export interface CrazyEightsConfig {
  cpuDifficulty: number;
  pointLimit: number;
}

/** Full Crazy Eights game state returned from the API. */
/** サーバー計算の推奨手。 */
export interface CrazyEightsHint {
  /** 推奨する手札の位置。スート選択フェーズでは undefined。 */
  cardIndex?: number;
  /** 8 を出した後に指名すべきスート。プレイフェーズでは undefined。 */
  suit?: number;
  /** 理由キー。 */
  reason: string;
}

export interface CrazyEightsResponse extends BaseGameResponse {
  /**
   * サーバー計算の推奨手。`hint` コマンドの応答にのみ載る。
   *
   * Hearts / Spades と同じく、理由付きの具体的な推奨を返す。フロント完結の
   * `FrontendHintTooltip` は全ゲーム共通の簡易ヒューリスティックで別物。
   */
  hint?: CrazyEightsHint;
  players: CrazyEightsPlayerData[];
  phase: number;
  roundNumber: number;
  currentPlayerIdx: number;
  discardTop: Card | null;
  drawPileCount: number;
  chosenSuit: number;
  gameEndFlag: boolean;
  winnerIdx: number;
  config: CrazyEightsConfig;
}

// --- Prší (プルシー) ---
