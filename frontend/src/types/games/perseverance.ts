// Type declarations for perseverance. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** A single tableau card in Perseverance. */
export interface PerseveranceTableauCard {
  card: Card | null;
  faceUp: boolean;
}

/** A suggested move hint in Perseverance. */
export interface PerseveranceHint {
  fromCol: number;
  cardIndex: number;
  toZone: string;
  toCol: number;
}

/** Full Perseverance game state returned from the API. */
export interface PerseveranceResponse extends BaseGameResponse {
  tableau: PerseveranceTableauCard[][];
  /** Card indices that start a movable descending same-suit run in each tableau column. */
  runStarts: number[][];
  foundation: Card[][];
  phase: number;
  moveCount: number;
  canUndo: boolean;
  isStalemate: boolean;
  /** Remaining gather-and-redeal chances (0–2). */
  redealsLeft: number;
  /** Legal destinations keyed by `tableau-<column>-<cardIndex>` source. */
  legalTargets: Record<string, { tableau: number[]; foundation: number[] }>;
  undoToEscape?: number;
  hint?: PerseveranceHint;
}

/** Source or target zone for a Perseverance card move. */
export interface PerseveranceMoveZone {
  zone: string;
  col?: number;
  cardIndex?: number;
}
