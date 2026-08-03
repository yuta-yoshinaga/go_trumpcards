// Type declarations for bisley. Split-file layout introduced by issue #4366;
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** A single tableau card in Bisley. */
export interface BisleyTableauCard {
  card: Card | null;
  faceUp: boolean;
}

/** A suggested move hint in Bisley. `toZone` is `'ace'`, `'king'`, or `'tableau'`. */
export interface BisleyHint {
  fromCol: number;
  toZone: string;
  toIdx: number;
}

/** Full Bisley game state returned from the API. */
export interface BisleyResponse extends BaseGameResponse {
  tableau: BisleyTableauCard[][];
  /** Ascending foundations, one per suit, built A -> K. */
  aceFoundations: Card[][];
  /** Descending foundations, one per suit, built K -> A; empty until its King is played. */
  kingFoundations: Card[][];
  phase: number;
  moveCount: number;
  canUndo: boolean;
  isStalemate: boolean;
  undoToEscape?: number;
  hint?: BisleyHint;
}

/** Source or target zone for a Bisley card move. */
export interface BisleyMoveZone {
  zone: string;
  col?: number;
}
