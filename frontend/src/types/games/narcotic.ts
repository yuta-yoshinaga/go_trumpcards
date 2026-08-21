// Type declarations for narcotic. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** A card in a Narcotic pile with action availability flags. */
export interface NarcoticCard {
  card: Card;
  top: boolean;
  /**
   * **盤面全体の性質。**露出4枚のランクが揃ったときだけ真で、そのときは4列とも真。
   * クローン元の Aces Up は列ごとに違ったが、Narcotic は4枚まとめてしか捨てない。
   */
  removable: boolean;
  /** この山の露出札を、同ランクを露出する左の山へ重ねられるか。 */
  movable: boolean;
}

/** A suggested hint in Narcotic. */
export interface NarcoticHint {
  /** `remove` and `redeal` carry no column, so `col` is -1 for those. */
  type: 'remove' | 'move' | 'draw' | 'redeal';
  col: number;
}

/** Full Narcotic game state returned from the API. */
export interface NarcoticResponse extends BaseGameResponse {
  /** The four piles. Only the last card of each is in play. */
  columns: NarcoticCard[][];
  stockCount: number;
  discardCount: number;
  /** The most recently discarded card; absent when nothing has been discarded. */
  discardTop?: Card | null;
  /** How many times the table has been gathered and re-dealt. **Unbounded.** */
  redealCount: number;
  phase: number;
  moveCount: number;
  canUndo: boolean;
  /** True when the same board has recurred with no legal move — the loss condition. */
  isStalemate: boolean;
  undoToEscape?: number;
  hint?: NarcoticHint;
}

// --- Pig's Tail ---
