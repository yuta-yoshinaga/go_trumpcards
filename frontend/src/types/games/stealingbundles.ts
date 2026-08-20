// Type declarations for stealingbundles. Split-file layout introduced by issue
// #4366; card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** One seat at a Stealing Bundles table. */
export interface StealingBundlesPlayer {
  id: number;
  isHuman: boolean;
  cardCount: number;
  /** Populated for the human player only; empty for the CPU. */
  cards: Card[];
  /** Cards won so far. **This is the score.** */
  bundleSize: number;
  /**
   * The top card of this seat's bundle, absent while the bundle is empty.
   *
   * **Public for every seat, including the CPUs.** It is the one thing
   * opponents aim at, so hiding it would hide the game.
   */
  bundleTop?: Card;
}

/** A suggestion. `victimIdx` is `-1` unless the advice is to steal. */
export interface StealingBundlesHint {
  cardIndex?: number;
  victimIdx: number;
  /** `stealingbundlesSteal`, `stealingbundlesTake`, or `stealingbundlesTrail`. */
  reason: string;
}

/** Table-size setting. */
export interface StealingBundlesConfig {
  /** Players at the table (2..4, default 4). */
  playerCnt: number;
}

/** Full Stealing Bundles game state returned from the API. */
export interface StealingBundlesResponse extends BaseGameResponse {
  players: StealingBundlesPlayer[];
  /** `0` = Play, `1` = GameEnd. */
  phase: number;
  tableCards: Card[];
  /**
   * Hand index (as a string key) to the table positions it can capture.
   *
   * Absent keys can capture nothing from the table.
   */
  tableMatches: Record<string, number[]>;
  /** Hand index (as a string key) to the seats whose bundle it can steal. */
  stealTargets: Record<string, number[]>;
  /**
   * Whether any capture is available.
   *
   * **You may only trail when this is false** — taking is compulsory when
   * something can be taken.
   */
  canCapture: boolean;
  deckRemaining: number;
  /** The seat that captured last, or `-1`. It receives whatever is left on the table. */
  lastCaptureIdx: number;
  /**
   * Whether that capture was a plain table take (`"take"`) or a whole bundle
   * stolen off another seat (`"steal"`); `""` before anyone has captured.
   *
   * A steal empties someone else's bundle, so it is not the same event as a take.
   */
  lastCaptureKind: string;
  /** The seat robbed by the last steal, or `-1` when the last capture was a take. */
  lastCaptureVictimIdx: number;
  currentPlayerIdx: number;
  turnNumber: number;
  packsDealt: number;
  gameEndFlag: boolean;
  /** `-1` until decided. */
  winnerIdx: number;
  hint?: StealingBundlesHint;
  config: StealingBundlesConfig;
}
