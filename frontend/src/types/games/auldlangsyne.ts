// Type declarations for auldlangsyne. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** A suggested move hint in Auld Lang Syne. */
export interface AuldLangSyneHint {
  wasteIdx: number;
  foundationIdx: number;
}

/**
 * Full Auld Lang Syne game state returned from the API.
 *
 * There is no `stockTop`, unlike SirTommyResponse: the deal is forced onto all
 * four wastes at once, so the player never sees the next card before it lands.
 */
export interface AuldLangSyneResponse extends BaseGameResponse {
  foundations: Card[][];
  wastes: Card[][];
  stockCount: number;
  phase: number;
  moveCount: number;
  canUndo: boolean;
  isStalemate: boolean;
  undoToEscape?: number;
  hint?: AuldLangSyneHint;
}

/** Source or target zone for an Auld Lang Syne card move. */
export interface AuldLangSyneMoveZone {
  zone: 'waste' | 'foundation';
  idx?: number;
}
