// Type declarations for tripeaks. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** A card in the TriPeaks tableau with removal and exposure status. */
export interface TriPeaksCard {
  card: Card | null;
  removed: boolean;
  exposed: boolean;
}

/** A suggested hint in TriPeaks. */
export interface TriPeaksHint {
  type: string;
  row: number;
  col: number;
}

/** Full TriPeaks game state returned from the API. */
export interface TriPeaksResponse extends BaseGameResponse {
  layout: TriPeaksCard[][];
  stockCount: number;
  waste: Card[];
  phase: number;
  moveCount: number;
  /**
   * Chain-bonus score, counted by the domain. The frontend used to derive this
   * itself, which left the same rule absent from the server and therefore out of
   * the CUI's reach entirely (#5511).
   */
  score: number;
  /** Length of the unbroken removal chain; 0 after a draw or an undo. */
  combo: number;
  canUndo: boolean;
  isStalemate: boolean;
  undoToEscape?: number;
  hint?: TriPeaksHint;
}
