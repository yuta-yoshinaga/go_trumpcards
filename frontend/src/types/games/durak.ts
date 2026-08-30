// Type declarations for durak. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Durak player data. */
export interface DurakPlayerData {
  id: number;
  isHuman: boolean;
  isFinished: boolean;
  cardCount: number;
  cards: Card[];
}

/** Durak table pair (attack + optional defense card). */
export interface DurakTablePair {
  attack: Card;
  defense: Card | null;
}

/** Durak CPU/human action record. */
export interface DurakAction {
  playerIdx: number;
  actionType: number; // 0=attack, 1=defend, 2=pass, 3=take, 4=transfer
  card: Card | null;
  attackIdx: number;
}

/** Durak game rule configuration. */
export interface DurakConfig {
  playerCount: number;
  cpuDifficulty: number;
  transferEnabled: boolean;
}

/** Input type alias for Durak configuration. */
export type DurakConfigInput = DurakConfig;

/** Full Durak game state returned from the API. */
/** サーバー計算の推奨手。 */
export interface DurakHint {
  /** 推奨する手札の位置。取る/パスを勧めるときは undefined。 */
  cardIndex?: number;
  /** 防御時に狙うテーブル上の攻撃カードの位置。 */
  attackIdx?: number;
  /** true なら「引き取る」を勧める。 */
  takeCards?: boolean;
  /** 理由キー。 */
  reason: string;
}

export interface DurakResponse extends BaseGameResponse {
  /**
   * サーバー計算の推奨手。`hint` コマンドの応答にのみ載る。
   *
   * フロント完結の `getDurakHint` は全ゲーム共通の簡易ヒューリスティックで別物。
   */
  hint?: DurakHint;
  players: DurakPlayerData[];
  currentTurn: number;
  phase: number;
  attackerIdx: number;
  defenderIdx: number;
  tablePairs: DurakTablePair[];
  trumpSuit: string;
  trumpCard: Card | null;
  stockCount: number;
  loserIdx: number;
  gameEndFlag: boolean;
  config: DurakConfig;
  cpuActions: DurakAction[];
  humanAction: DurakAction | null;
  boutNumber: number;
  sortMode: number;
}

// --- Forty Thieves (フォーティシーブス) ---
