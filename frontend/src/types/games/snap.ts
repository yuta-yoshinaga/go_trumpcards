// Type declarations for snap. Split-file layout introduced by issue #4366;
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** One seat at a Snap table. Stocks are face down — only the count is public. */
export interface SnapPlayer {
  id: number;
  isHuman: boolean;
  stockSize: number;
}

/** A suggestion: whether to call right now, and why. */
export interface SnapHint {
  /** `true` when a call would be correct this instant. */
  snap: boolean;
  /** `snapDeclare`, `snapStep`, or `snapWait`. */
  reason: string;
}

/** Full Snap game state returned from the API. */
export interface SnapResponse extends BaseGameResponse {
  /** `0` = Play, `1` = GameEnd. */
  phase: number;
  gameEndFlag: boolean;
  /** `-1` until decided, and `-1` when play simply could not continue. */
  winnerIdx: number;
  currentTurnIdx: number;
  isHumanTurn: boolean;
  /**
   * Whether a call would be correct right now. **True only when the top two
   * cards match** — a single card showing can never be snapped, because there
   * is no previous card to match it against.
   */
  snapAvailable: boolean;
  centerPileSize: number;
  topCard?: Card;
  players: SnapPlayer[];
  playerCnt: number;
  /** `0` = easy, `1` = normal, `2` = hard. Drives the CPU's reaction time. */
  cpuDifficulty: number;
  /** `0` = none, `1` = a CPU has booked a call, `2` = a CPU will turn a card. */
  pendingKind: number;
  /** When the booked action fires. A human who calls before this wins the pile. */
  pendingDeadlineMs: number;
  /** `0` none, `1` step, `2` correct call, `3` wrong call, `4` out of stock. */
  lastEventKind: number;
  lastEventPlayerIdx: number;
  hint?: SnapHint;
}
