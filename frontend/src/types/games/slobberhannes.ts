// Type declarations for slobberhannes. Split-file layout introduced by issue
// #4366; card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** A card played into the current Slobberhannes trick. */
export interface SlobberhannesTrickCard {
  playerIdx: number;
  card: Card;
}

/** One seat at a Slobberhannes table. */
export interface SlobberhannesPlayer {
  id: number;
  isHuman: boolean;
  /** Hand size. The only hand information exposed for the CPU. */
  cardCount: number;
  /** Populated for the human player only; empty for the CPU. */
  cards: Card[];
  /** Running total across rounds. Penalties are negative, the clean-round bonus positive. */
  score: number;
  trickCount: number;
  /** Took this round's first trick (−1). */
  tookFirstTrick: boolean;
  /** Took this round's last trick (−1). */
  tookLastTrick: boolean;
  /** Took the trick holding the Q of clubs (−1). */
  tookQueen: boolean;
}

/** A suggested card to play. */
export interface SlobberhannesHint {
  cardIndex?: number;
  /**
   * Why that card: `slobberhannesAvoid` (this trick carries a penalty),
   * `slobberhannesDump` (safe trick — shed a high card now), or
   * `slobberhannesLeadLow` (leading; play low and see what happens).
   */
  reason: string;
}

/** Round-count setting. */
export interface SlobberhannesConfig {
  /** Rounds to play (1..12, default 4). */
  rounds: number;
}

/** Full Slobberhannes game state returned from the API. */
export interface SlobberhannesResponse extends BaseGameResponse {
  players: SlobberhannesPlayer[];
  /** `0` = Play, `1` = RoundEnd, `2` = GameEnd. */
  phase: number;
  roundNumber: number;
  /** 0-based. Tricks 0 and 7 carry the position penalties. */
  trickNumber: number;
  currentPlayerIdx: number;
  leadPlayerIdx: number;
  dealerIdx: number;
  currentTrick: SlobberhannesTrickCard[];
  /** Hand indices you may legally play. Following suit is compulsory. */
  validPlays: number[];
  gameEndFlag: boolean;
  /** `-1` until decided, and `-1` on a tie. */
  winnerIdx: number;
  hint?: SlobberhannesHint;
  config: SlobberhannesConfig;
}
