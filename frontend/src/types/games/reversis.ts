// Type declarations for reversis. Split-file layout introduced by issue
// #4366; card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** A card played into the current Reversis trick. */
export interface ReversisTrickCard {
  playerIdx: number;
  card: Card;
}

/** One seat at a Reversis table. */
export interface ReversisPlayer {
  id: number;
  isHuman: boolean;
  /** Hand size. The only hand information exposed for the CPU. */
  cardCount: number;
  /** Populated for the human player only; empty for the CPU. */
  cards: Card[];
  /** Chips held. **Higher is better** — this decides the game. */
  chips: number;
  /** Penalty points taken this round. **Lower is better** — it wins the pool. */
  roundPenalty: number;
  trickCount: number;
  /** Took the J of hearts (+5 penalty, 5 chips to the pool). */
  tookQuinola: boolean;
  /** Took the A of diamonds (+5 penalty, 5 chips to the pool). */
  tookDiamondAce: boolean;
}

/** A suggested card to play. */
export interface ReversisHint {
  cardIndex?: number;
  /**
   * Why that card: `reversisAvoidMarked` (the Quinola or the A of diamonds is
   * on the table), `reversisAvoidPoints` (plain penalty cards are),
   * `reversisLeadSafe` (leading; play a low card worth nothing), or
   * `reversisDumpHigh` (safe trick — shed a high card now).
   */
  reason: string;
}

/** Round-count setting. */
export interface ReversisConfig {
  /** Rounds to play (1..12, default 4). */
  rounds: number;
}

/** Full Reversis game state returned from the API. */
export interface ReversisResponse extends BaseGameResponse {
  players: ReversisPlayer[];
  /** `0` = Play, `1` = RoundEnd, `2` = GameEnd. */
  phase: number;
  roundNumber: number;
  trickNumber: number;
  /** Chips in the pool, taken whole by the fewest penalty points. */
  pool: number;
  currentPlayerIdx: number;
  leadPlayerIdx: number;
  dealerIdx: number;
  currentTrick: ReversisTrickCard[];
  /** Hand indices you may legally play. Following suit is compulsory. */
  validPlays: number[];
  gameEndFlag: boolean;
  /** `-1` until decided, and `-1` on a tie. */
  winnerIdx: number;
  hint?: ReversisHint;
  config: ReversisConfig;
}
