// Type declarations for sergeantmajor. Split-file layout introduced by issue
// #4366; card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** A card played into the current Sergeant Major trick. */
export interface SergeantMajorTrickCard {
  playerIdx: number;
  card: Card;
}

/** One seat at a Sergeant Major table. */
export interface SergeantMajorPlayer {
  id: number;
  isHuman: boolean;
  /** Hand size. The only hand information exposed for the CPU. */
  cardCount: number;
  /** Populated for the human player only; empty for the CPU. */
  cards: Card[];
  /**
   * Tricks this seat owes this round: `8`, `5` or `3`. **Fixed by seat, never
   * bid** — the dealer owes 8, the seat to their left 5, the next 3.
   */
  target: number;
  trickCount: number;
  /** Running total of (tricks − target). **The game is decided on this.** */
  score: number;
}

/**
 * A suggestion. While choosing trump it names a suit; while discarding it
 * names several cards in `indices`; during play it names one card.
 */
export interface SergeantMajorHint {
  cardIndex?: number;
  /**
   * `sergeantmajorSelectTrump` before play; `sergeantmajorDiscard` while
   * sorting the kitty; `sergeantmajorWinTrick` or `sergeantmajorPressOn`
   * (target already met) during play.
   */
  reason: string;
  /** Suit to make trump. `0` otherwise. */
  suit: number;
  /** Cards to discard. Empty outside the discard phase. */
  indices: number[];
}

/** Round-count setting. */
export interface SergeantMajorConfig {
  /** Rounds to play (3..30, default 3 — one turn each at 8, 5 and 3). */
  rounds: number;
}

/** Full Sergeant Major game state returned from the API. */
export interface SergeantMajorResponse extends BaseGameResponse {
  players: SergeantMajorPlayer[];
  /** `0` = Trump, `1` = Discard, `2` = Play, `3` = RoundEnd, `4` = GameEnd. */
  phase: number;
  roundNumber: number;
  trickNumber: number;
  /** `0` until the dealer has chosen. */
  trumpSuit: number;
  /**
   * Cards still face down. 52 does not divide by three, so the four left over
   * go to the dealer, who discards `discardCount` after naming trump.
   */
  kittySize: number;
  /**
   * Positions in the human's hand that came in from the kitty.
   *
   * Absorbing the kitty mixes four new cards into a re-sorted hand, so nothing
   * distinguishes them while the dealer picks four to throw away (#5759). Empty
   * once the discard is done.
   */
  kittyIndices?: number[];
  /** How many cards the dealer must discard (4). */
  discardCount: number;
  /**
   * How many cards changed hands for last round's shortfall. Falling short of
   * your target costs your best cards; nothing on the board shows this.
   */
  lastExchange: number;
  currentPlayerIdx: number;
  leadPlayerIdx: number;
  /** The dealer. **This seat owes 8 and names trump.** */
  dealerIdx: number;
  currentTrick: SergeantMajorTrickCard[];
  /** Hand indices you may legally play. Following suit is compulsory. */
  validPlays: number[];
  gameEndFlag: boolean;
  /** `-1` until decided, and `-1` on a tie. */
  winnerIdx: number;
  hint?: SergeantMajorHint;
  config: SergeantMajorConfig;
}
