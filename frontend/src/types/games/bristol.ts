// Type declarations for bristol. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** A suggested move hint in Bristol. */
export interface BristolHint {
  fromZone: string;
  fromCol: number;
  toZone: string;
  toCol: number;
}

/** Full Bristol game state returned from the API. */
export interface BristolResponse extends BaseGameResponse {
  tableau: Card[][];
  fan: Card[][];
  stockCount: number;
  foundation: Card[][];
  /**
   * 移動元ごとの合法な移動先。キーは `"tableau-0"` / `"fan-2"`。
   *
   * 選択中は全ての移動先が同じ見た目で強調されていて、押すまで合法か
   * 分からなかった (#4813)。
   */
  legalTargets: Record<string, { tableau: number[]; foundation: number[] }>;
  phase: number;
  moveCount: number;
  canUndo: boolean;
  hint?: BristolHint;
}

/** Source or target zone for a Bristol card move. */
export interface BristolMoveZone {
  zone: string;
  col?: number;
}

// --- La Belle Lucie (ラ・ベル・ルーシー) ---
