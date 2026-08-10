// Type declarations for tarneeb. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Tarneeb player data with team and current bid. */
export interface TarneebPlayerData {
  id: number;
  isHuman: boolean;
  team: number;
  cardCount: number;
  cards: Card[];
  bid: number;
  roundScore: number;
  cumulativeScore: number;
  trickCount: number;
}

/** A card played in a Tarneeb trick. */
export interface TarneebTrickCard {
  playerIdx: number;
  card: Card;
}

/** Tarneeb game configuration. */
export interface TarneebConfig {
  cpuDifficulty: number;
  pointLimit: number;
  minBid: number;
}

/** A suggested hint for Tarneeb. */
export interface TarneebHint {
  cardIndex?: number;
  bid?: number;
  trumpSuit?: number;
  reason: string;
}

/** Full Tarneeb game state returned from the API. */
export interface TarneebResponse extends BaseGameResponse {
  players: TarneebPlayerData[];
  teamScores: number[];
  phase: number;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  bidPlayerIdx: number;
  bidWinnerIdx: number;
  highestBid: number;
  trumpSuit: number;
  redealCount: number;
  dealerIdx: number;
  currentTrick: TarneebTrickCard[];
  gameEndFlag: boolean;
  winnerTeam: number;
  leadPlayerIdx: number;
  /**
   * 人間がいま出せる手札の位置。
   *
   * マストフォローの判定はドメインの `GetValidPlayIndices` が持っており、
   * フロントで組み立て直すと片方だけ直したときに黙って食い違う (#4713)。
   * プレイフェーズで人間の手番でなければ空 — **空を「制限なし」と読まないこと**。
   */
  validPlayIndices: number[];
  config: TarneebConfig;
  hint?: TarneebHint;
}

// --- High Card Flush (ハイカードフラッシュ) ---
