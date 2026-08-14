// Type declarations for bhabhi. Split-file layout introduced by issue
// #4366; card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** A card sitting on the table in the current Bhabhi trick. */
export interface BhabhiTrickCard {
  playerIdx: number;
  card: Card;
}

/** One seat at a Bhabhi table. */
export interface BhabhiPlayer {
  id: number;
  isHuman: boolean;
  /** Hand size. The only hand information exposed for the CPU. */
  cardCount: number;
  /** Populated for the human player only; empty for the CPU. */
  cards: Card[];
  /**
   * Order in which this seat emptied its hand, or `-1` while still holding
   * cards. **This is not a ranking by skill** — everyone who gets out is
   * simply safe, and only the last one still holding cards loses.
   */
  rank: number;
  /** How many times this seat has taken the pile into hand. */
  pickups: number;
}

/** A suggestion naming a card to play. */
export interface BhabhiHint {
  cardIndex?: number;
  /**
   * `bhabhiLead` when starting a trick, `bhabhiDuck` when you can follow, and
   * `bhabhiDumpHigh` when you cannot and will take the pile regardless.
   */
  reason: string;
}

/** Table-size setting. */
export interface BhabhiConfig {
  /** How many seats play (3..7, default 4). */
  playerCnt: number;
}

/** Full Bhabhi game state returned from the API. */
export interface BhabhiResponse extends BaseGameResponse {
  players: BhabhiPlayer[];
  /** `0` = Play, `1` = GameEnd. There are no hands to split the game into. */
  phase: number;
  /** How many tricks have closed, by either a completed trick or a pickup. */
  trickNumber: number;
  /** `0` before anyone has led. Following this suit is compulsory. */
  leadSuit: number;
  /** Cards on the table. **Whoever cannot follow takes all of them.** */
  pile: BhabhiTrickCard[];
  /** Seat that last took the pile, or `-1` before the first pickup. */
  lastPickupIdx: number;
  /** How many cards that pickup was worth. */
  lastPickupSize: number;
  currentPlayerIdx: number;
  leadPlayerIdx: number;
  /** Hand indices you may legally play. Following suit is compulsory. */
  validPlays: number[];
  /** How many seats still hold cards. */
  aliveCount: number;
  gameEndFlag: boolean;
  /** The loser, or `-1` until decided. **This game names a loser, not a winner.** */
  bhabhiIdx: number;
  /**
   * The game was cut short as deadlocked rather than played out. Cards only
   * leave play on a trick everyone followed, so a distribution where somebody
   * is always void can circulate forever; at `stalemateTricks` the seat
   * holding the most cards becomes the Bhabhi.
   */
  stalemate: boolean;
  /** The trick count at which a deadlock is called. */
  stalemateTricks: number;
  hint?: BhabhiHint;
  config: BhabhiConfig;
}
