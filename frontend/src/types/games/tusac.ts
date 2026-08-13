// Type declarations for tusac. Split-file layout introduced by issue #4366;
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Phases, matching the Go domain. */
export const TUSAC_PHASE = { draw: 0, discard: 1, roundEnd: 2, gameEnd: 3 } as const;

/** Combination kinds, matching the Go domain. */
export const TUSAC_MELD = { none: 0, sameColorSet: 1, chariotTrio: 2, soldierSet: 3 } as const;

/**
 * Colours, as `design` values on the shared `Card`.
 *
 * **The deck is not a 52-card pack.** Suits are colours and ranks are
 * chess-piece names, so the usual suit symbols do not apply.
 */
export const TUSAC_COLOR = { yellow: 1, red: 2, green: 3, white: 4 } as const;

/** Piece types, as `value` on the shared `Card`. */
export const TUSAC_PIECE = {
  general: 1,
  advisor: 2,
  elephant: 3,
  chariot: 4,
  horse: 5,
  cannon: 6,
  soldier: 7,
} as const;

/** A suggestion for the current decision. */
export interface TuSacHint {
  /** `draw`, `meld`, `discard` or `next`. */
  action: string;
  /** 0-based hand positions the hint points at. */
  indexes: number[];
  reason: string;
}

/** One combination laid on the table. */
export interface TuSacMeld {
  /** 1=three of one colour, 2=chariot-horse-cannon, 3=five soldiers. */
  kind: number;
  points: number;
  cards: Card[];
}

/** One seat at the table. */
export interface TuSacSeat {
  name: string;
  isHuman: boolean;
  /**
   * The seat's hand.
   *
   * **Empty for every seat but the human.** Other hands are never put on the
   * wire — this game is read from the melds on the table, not from anyone's
   * hand.
   */
  cards: Card[];
  /** How many cards the seat holds. Known for every seat. */
  handCount: number;
  /** Combinations laid on the table. **Visible for every seat.** */
  melds: TuSacMeld[];
  meldPoints: number;
  /** Running total across rounds. */
  score: number;
  /** Melds minus cards still held, for the round just scored. */
  roundScore: number;
  isTurn: boolean;
  wentOut: boolean;
}

/** Tu Sac game settings. */
export interface TuSacConfig {
  seats: number;
  rounds: number;
}

/** Response payload for `/tusac/exec`. */
export interface TuSacResponse extends BaseGameResponse {
  /** 0=Draw, 1=Discard, 2=RoundEnd, 3=GameEnd. */
  phase: number;
  seats: TuSacSeat[];
  /** The top of the discard pile, or null when it is empty. */
  discardTop: Card | null;
  discardCount: number;
  stockCount: number;
  turnSeat: number;
  humanSeat: number;
  isHumanTurn: boolean;
  roundNumber: number;
  rounds: number;
  /** The seat that melded out, or -1 when the stock ran out instead. */
  wentOutSeat: number;
  /** Cards dealt to each seat (20). */
  handSize: number;
  /** Always 112 — four colours, seven piece types, four copies. */
  deckSize: number;
  /**
   * Score for each combination kind, indexed by kind.
   *
   * **Sent by the server so the page never hardcodes them** — a copy here
   * would drift from the odds the scores were derived from.
   */
  meldPointsByKind: number[];
  winnerSeat: number;
  gameEndFlag: boolean;
  config?: TuSacConfig;
}
