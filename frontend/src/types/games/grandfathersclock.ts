// Type declarations for grandfathersclock. Split-file layout introduced by
// issue #4366; card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** A single tableau card in Grandfather's Clock. */
export interface GrandfathersClockTableauCard {
  card: Card | null;
  faceUp: boolean;
}

/** One clock face: its cards plus the rank it has to end on. */
export interface GrandfathersClockFoundation {
  cards: Card[];
  /** The rank this face must reach — 1 at one o'clock, 12 at twelve. */
  targetRank: number;
  /** Whether the face has reached `targetRank` and accepts nothing more. */
  complete: boolean;
}

/** A suggested move hint in Grandfather's Clock. */
export interface GrandfathersClockHint {
  fromCol: number;
  /** `'foundation'` or `'tableau'`. */
  toZone: string;
  /** Destination clock face or column index. */
  toIdx: number;
}

/** Full Grandfather's Clock game state returned from the API. */
export interface GrandfathersClockResponse extends BaseGameResponse {
  tableau: GrandfathersClockTableauCard[][];
  /** Twelve clock faces, index 0 at one o'clock through index 11 at twelve. */
  foundation: GrandfathersClockFoundation[];
  phase: number;
  moveCount: number;
  canUndo: boolean;
  isStalemate: boolean;
  undoToEscape?: number;
  hint?: GrandfathersClockHint;
}

/** Source or target zone for a Grandfather's Clock card move. */
export interface GrandfathersClockMoveZone {
  zone: string;
  col?: number;
}
