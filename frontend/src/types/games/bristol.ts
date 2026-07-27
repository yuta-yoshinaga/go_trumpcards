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
