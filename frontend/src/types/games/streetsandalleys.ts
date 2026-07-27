// Type declarations for streetsandalleys. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** A single tableau card in Streets and Alleys. */
export interface StreetsAndAlleysTableauCard {
  card: Card | null;
  faceUp: boolean;
}

/** A suggested move hint in Streets and Alleys. */
export interface StreetsAndAlleysHint {
  fromCol: number;
  cardIndex: number;
  toZone: string;
  toCol: number;
}

/** Full Streets and Alleys game state returned from the API. */
export interface StreetsAndAlleysResponse extends BaseGameResponse {
  tableau: StreetsAndAlleysTableauCard[][];
  foundation: Card[][];
  phase: number;
  moveCount: number;
  canUndo: boolean;
  isStalemate: boolean;
  undoToEscape?: number;
  hint?: StreetsAndAlleysHint;
}

/** Source or target zone for a Streets and Alleys card move. */
export interface StreetsAndAlleysMoveZone {
  zone: string;
  col?: number;
  cardIndex?: number;
}

// --- King Albert ---
