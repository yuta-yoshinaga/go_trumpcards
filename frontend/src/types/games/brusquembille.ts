// Type declarations for brusquembille. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Brusquembille player data with points and trick count. */
export interface BrusquembillePlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  points: number;
  trickCount: number;
}

/** A card played in a Brusquembille trick. */
export interface BrusquembilleTrickCard {
  playerIdx: number;
  card: Card;
}

/** Brusquembille game configuration. */
export interface BrusquembilleConfig {
  cpuDifficulty: number;
}

/** A suggested hint for Brusquembille. */
export interface BrusquembilleHint {
  cardIndex?: number;
  reason: string;
}

/** Full Brusquembille game state returned from the API. */
export interface BrusquembilleResponse extends BaseGameResponse {
  players: BrusquembillePlayerData[];
  phase: number;
  trickNumber: number;
  currentPlayerIdx: number;
  currentTrick: BrusquembilleTrickCard[];
  trumpSuit: number;
  /** Face-up trump card (omitted once the stock is exhausted). */
  trumpCard?: Card;
  dealerIdx: number;
  leadPlayerIdx: number;
  /**
   * Cards remaining in the stock; this does NOT include the face-up trump
   * card (which is tracked separately via `trumpCard` until drawn last).
   */
  stockRemaining: number;
  gameEndFlag: boolean;
  /** -1 = tie or unfinished. */
  winnerIdx: number;
  /**
   * Hand positions the human may legally play right now.
   *
   * While the stock lasts this is every card; once the stock and the face-up
   * trump are gone, following the led suit is compulsory and this narrows to
   * the cards that follow. **The page must not re-derive this** — the backend
   * owns the rule, and two copies drift apart.
   */
  validIndices: number[];
  /** True once following the led suit is compulsory (the stock has run out). */
  followRequired: boolean;
  config: BrusquembilleConfig;
  hint?: BrusquembilleHint;
}

// --- Schnapsen / Sixty-Six (シュナプセン / 66) ---
