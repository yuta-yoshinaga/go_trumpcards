// Type declarations for lingerlonger. Split-file layout introduced by issue
// #4366; card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** A card played into the current Linger Longer trick. */
export interface LingerLongerTrickCard {
  playerIdx: number;
  card: Card;
}

/** One seat at a Linger Longer table. */
export interface LingerLongerPlayer {
  id: number;
  isHuman: boolean;
  /** Hand size. **More is better** — running out is how you go out. */
  cardCount: number;
  /** Populated for the human player only; empty for the CPU. */
  cards: Card[];
  /** Tricks won. **Not a score** — it is the number of refills earned. */
  tricksWon: number;
  /** Elimination order, or `0` while still holding cards. */
  eliminatedAt: number;
}

/** A suggestion. Always names a card; there is only ever one kind of move. */
export interface LingerLongerHint {
  cardIndex?: number;
  /** `lingerlongerWinTrick`, `lingerlongerNoStock`, or `lingerlongerDuck`. */
  reason: string;
}

/** Table-size setting. */
export interface LingerLongerConfig {
  /** Players at the table (3..6, default 4). You are dealt that many cards. */
  playerCnt: number;
}

/** Full Linger Longer game state returned from the API. */
export interface LingerLongerResponse extends BaseGameResponse {
  players: LingerLongerPlayer[];
  /** `0` = Play, `1` = GameEnd. */
  phase: number;
  /** Hand indices you may legally play. Following suit is compulsory. */
  validPlays: number[];
  /**
   * Cards left in the stock.
   *
   * **At `0` nobody can refill any more**, so from here every trick strictly
   * shrinks the table and eliminations come quickly.
   */
  stockSize: number;
  currentTrick: LingerLongerTrickCard[];
  currentPlayerIdx: number;
  leadPlayerIdx: number;
  /** Tricks resolved so far. */
  trickNumber: number;
  /** The seat that just drew a replacement, or `-1`. */
  lastDrawIdx: number;
  /** How many seats have run out so far. */
  eliminatedCnt: number;
  /** Cards out of play. **Only a completed trick removes any.** */
  discarded: number;
  gameEndFlag: boolean;
  /** `-1` until decided. */
  winnerIdx: number;
  hint?: LingerLongerHint;
  config: LingerLongerConfig;
}
