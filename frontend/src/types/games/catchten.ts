// Type declarations for catchten. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Catch the Ten player data with team, scores, and trick count. */
export interface CatchTenPlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  roundScore: number;
  cumulativeScore: number;
  trickCount: number;
  team: number;
}

/** A card played in a Catch the Ten trick. */
export interface CatchTenTrickCard {
  playerIdx: number;
  card: Card;
}

/** Catch the Ten game configuration. */
export interface CatchTenConfig {
  cpuDifficulty: number;
  pointLimit: number;
}

/** A suggested hint for Catch the Ten. */
export interface CatchTenHint {
  cardIndex?: number;
  reason: string;
}

/** Full Catch the Ten game state returned from the API. */
export interface CatchTenResponse extends BaseGameResponse {
  players: CatchTenPlayerData[];
  phase: number;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  currentTrick: CatchTenTrickCard[];
  trumpSuit: number;
  dealerIdx: number;
  teamScores: [number, number];
  gameEndFlag: boolean;
  winnerTeam: number;
  leadPlayerIdx: number;
  /**
   * 人間がいま出せる手札の位置。
   *
   * フォロースートの判定はドメインの `GetValidPlayIndices` が持っており、
   * フロントで組み立て直すと片方だけ直したときに黙って食い違う。
   * プレイフェーズで人間の手番でなければ空 — **空を「制限なし」と読まないこと**。
   */
  validPlayIndices: number[];
  config: CatchTenConfig;
  hint?: CatchTenHint;
}

// --- Briscola (ブリスコラ) ---
