// Type declarations for julepe. Split-file layout introduced by issue
// #4366; card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** A card played into the current Julepe trick. */
export interface JulepeTrickCard {
  playerIdx: number;
  card: Card;
}

/** One seat at a Julepe table. */
export interface JulepePlayer {
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
export interface JulepeHint {
  cardIndex?: number;
  /**
   * `julepePlayIn` / `julepePassOut` in the decision phase; `julepeTakeTrick` (no
   * trick banked yet, so avoid the extra payment) or `julepeAlreadySafe` (one
   * trick banked) during play.
   */
  reason: string;
}

/** Table-size and round-count settings. */
export interface JulepeConfig {
  /** Players at the table (3..5, default 4). */
  playerCnt: number;
  /** Rounds to play (1..12, default 4). */
  rounds: number;
}

/** Full Julepe game state returned from the API. */
export interface JulepeResponse extends BaseGameResponse {
  /** Length is 3, 4 or 5 depending on the configured table size. */
  players: JulepePlayer[];
  /** `0` = Decide, `1` = Play, `2` = RoundEnd, `3` = GameEnd. */
  phase: number;
  roundNumber: number;
  trickNumber: number;
  /** Chips in the pot, shared out by tricks taken. */
  pot: number;
  /** 現在の参加人数に対する規定トリック数。**人数で変わる**ので固定値にしない。 */
  requiredTricks: number;
  /** 次ラウンドのアンティが倍になる席。 */
  beast: boolean[];
  /** Suit of the card turned after the deal. */
  trumpSuit: number;
  /** The card turned to fix trump. */
  upCard?: Card;
  currentPlayerIdx: number;
  leadPlayerIdx: number;
  dealerIdx: number;
  /** How many players entered this round. */
  activeCount: number;
  currentTrick: JulepeTrickCard[];
  /** Hand indices you may legally play. Following suit is compulsory. */
  validPlays: number[];
  gameEndFlag: boolean;
  /** `-1` until decided, and `-1` on a tie. */
  winnerIdx: number;
  hint?: JulepeHint;
  config: JulepeConfig;
}
