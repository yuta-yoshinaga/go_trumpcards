// Type declarations for pasur. Split-file layout introduced by issue
// #4366; card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** One seat at a Pasur table. */
export interface PasurPlayer {
  id: number;
  isHuman: boolean;
  /** Hand size. The only hand information exposed for the CPU. */
  cardCount: number;
  /** Populated for the human player only; empty for the CPU. */
  cards: Card[];
  /** Cards taken so far, across both the normal and the soor pile. */
  capturedCount: number;
  /** How many times this seat emptied the table. Those captures score double. */
  soors: number;
  score: number;
}

/**
 * A suggestion. `table` names the table indices to take, and is empty when the
 * advice is to lay the card down instead.
 */
export interface PasurHint {
  cardIndex?: number;
  /** `pasurSoor`, `pasurCapture`, or `pasurTrail`. */
  reason: string;
  /** Table indices to take. Empty for a trail. */
  table: number[];
}

/** Table-size setting. */
export interface PasurConfig {
  /** Players at the table (2..4, default 4). */
  playerCnt: number;
}

/** Full Pasur game state returned from the API. */
export interface PasurResponse extends BaseGameResponse {
  players: PasurPlayer[];
  /** `0` = Play, `1` = GameEnd. */
  phase: number;
  /** The cards face up on the table. Their positions are the `table` indices. */
  table: Card[];
  /**
   * Per hand card, the sets of table indices it may take — so a client never
   * has to rebuild "which subsets add to 11", which would drift from the
   * server. An empty array for a card means it can only be laid down.
   */
  captureOptions: number[][][];
  deckRemaining: number;
  packsDealt: number;
  /** The seat that captured last. **Cards left on the table go to them.** */
  lastCaptureIdx: number;
  currentPlayerIdx: number;
  gameEndFlag: boolean;
  /** Seats on the highest score. More than one means a tie; empty until decided. */
  winners: number[];
  hint?: PasurHint;
  config: PasurConfig;
}
