// Type declarations for rollingstone. Split-file layout introduced by issue
// #4366; card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** A card played into the current Rolling Stone trick. */
export interface RollingStoneTrickCard {
  playerIdx: number;
  card: Card;
}

/** One seat at a Rolling Stone table. */
export interface RollingStonePlayer {
  id: number;
  isHuman: boolean;
  /** Hand size. **Fewer is better** — running out first is how you win. */
  cardCount: number;
  /** Populated for the human player only; empty for the CPU. */
  cards: Card[];
  /** How many tricks this seat has been forced to take into hand. */
  pickups: number;
  /** Finishing place, or `0` while still holding cards. */
  finishedAt: number;
}

/** A suggestion. Carries no card index when the only move is to pick up. */
export interface RollingStoneHint {
  cardIndex?: number;
  /** `rollingstoneLead`, `rollingstoneFollow`, or `rollingstonePickUp`. */
  reason: string;
}

/** Table-size setting. */
export interface RollingStoneConfig {
  /** Players at the table (4..6, default 4). The deck is sized to match. */
  playerCnt: number;
}

/** Full Rolling Stone game state returned from the API. */
export interface RollingStoneResponse extends BaseGameResponse {
  players: RollingStonePlayer[];
  /** `0` = Play, `1` = GameEnd. */
  phase: number;
  /**
   * Whether you cannot follow and must take the trick into your hand.
   *
   * **Read this before `validPlays`** — an empty `validPlays` also means "not
   * your turn", and the two call for opposite things.
   */
  mustPickUp: boolean;
  /** Hand indices you may legally play. Following suit is compulsory. */
  validPlays: number[];
  currentTrick: RollingStoneTrickCard[];
  currentPlayerIdx: number;
  leadPlayerIdx: number;
  /** Tricks resolved so far. Pickups count too — they end a trick as well. */
  trickNumber: number;
  /** The seat that just took a trick into hand, or `-1`. */
  lastPickupIdx: number;
  finishedCnt: number;
  /** Cards in this deal: players x 8, so 32 / 40 / 48. */
  deckSize: number;
  /** Cards out of play. **Only a completed trick removes any.** */
  discarded: number;
  gameEndFlag: boolean;
  /** `-1` until decided. */
  winnerIdx: number;
  hint?: RollingStoneHint;
  config: RollingStoneConfig;
}
