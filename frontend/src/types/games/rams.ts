// Type declarations for rams. Split-file layout introduced by issue
// #4366; card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** A card played into the current Rams trick. */
export interface RamsTrickCard {
  playerIdx: number;
  card: Card;
}

/** One seat at a Rams table. */
export interface RamsPlayer {
  id: number;
  isHuman: boolean;
  /** Hand size. The only hand information exposed for the CPU. */
  cardCount: number;
  /** Populated for the human player only; empty for the CPU. */
  cards: Card[];
  /** Chips held. **Higher is better** — this decides the game. */
  chips: number;
  /** Entered this round (chose to play rather than drop). */
  inRound: boolean;
  /** Has chosen play or drop for this round. */
  decided: boolean;
  /** Tricks taken this round. Zero while `inRound` costs an extra payment. */
  roundTricks: number;
  trickCount: number;
}

/**
 * A suggestion. In the decision phase it advises whether to enter and carries
 * no `cardIndex`; during play it names a card.
 */
export interface RamsHint {
  cardIndex?: number;
  /**
   * `ramsPlayIn` / `ramsPassOut` in the decision phase; `ramsTakeTrick` (no
   * trick banked yet, so avoid the extra payment) or `ramsAlreadySafe` (one
   * trick banked) during play.
   */
  reason: string;
}

/** Table-size and round-count settings. */
export interface RamsConfig {
  /** Players at the table (3..5, default 4). */
  playerCnt: number;
  /** Rounds to play (1..12, default 4). */
  rounds: number;
}

/** Full Rams game state returned from the API. */
export interface RamsResponse extends BaseGameResponse {
  /** Length is 3, 4 or 5 depending on the configured table size. */
  players: RamsPlayer[];
  /** `0` = Decide, `1` = Play, `2` = RoundEnd, `3` = GameEnd. */
  phase: number;
  roundNumber: number;
  trickNumber: number;
  /** Chips in the pot, shared out by tricks taken. */
  pot: number;
  /** Suit of the card turned after the deal. */
  trumpSuit: number;
  /** The card turned to fix trump. */
  upCard?: Card;
  currentPlayerIdx: number;
  leadPlayerIdx: number;
  dealerIdx: number;
  /** How many players entered this round. */
  activeCount: number;
  currentTrick: RamsTrickCard[];
  /** Hand indices you may legally play. Following suit is compulsory. */
  validPlays: number[];
  gameEndFlag: boolean;
  /** `-1` until decided, and `-1` on a tie. */
  winnerIdx: number;
  hint?: RamsHint;
  config: RamsConfig;
}
