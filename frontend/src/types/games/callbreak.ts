// Type declarations for callbreak. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Call Break player data with bid and integer×10 scores. */
export interface CallBreakPlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  bid: number;
  /** Round score in internal int×10 form (e.g. 41 == 4.1 points). */
  roundScore: number;
  /** Cumulative score in internal int×10 form. */
  cumulativeScore: number;
  trickCount: number;
  /**
   * ビッドを超えて取った余剰トリック数 (バッグ)。
   *
   * ページ側で `trickCount - bid` を組み立てると CUI と式が二重化して黙って
   * 食い違うので、ドメインの `GetBags()` の結果をそのまま受け取る (#4752)。
   */
  bags: number;
}

/** A card played in a Call Break trick. */
export interface CallBreakTrickCard {
  playerIdx: number;
  card: Card;
}

/** Call Break game configuration. */
export interface CallBreakConfig {
  cpuDifficulty: number;
  maxRounds: number;
}

/** A suggested hint for Call Break. */
export interface CallBreakHint {
  cardIndex?: number;
  bid?: number;
  reason: string;
}

/** Full Call Break game state returned from the API. */
export interface CallBreakResponse extends BaseGameResponse {
  players: CallBreakPlayerData[];
  phase: number;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  bidPlayerIdx: number;
  currentTrick: CallBreakTrickCard[];
  spadesBroken: boolean;
  gameEndFlag: boolean;
  winnerIdx: number;
  leadPlayerIdx: number;
  config: CallBreakConfig;
  hint?: CallBreakHint;
  /**
   * Indices in the human player's hand that are legal to play this turn.
   * Empty array outside the play phase / when it is not the human's turn.
   */
  validPlayIndices: number[];
}

// --- Pitch (Setback / Auction Pitch) ---
