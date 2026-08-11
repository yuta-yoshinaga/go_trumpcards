// Type declarations for germanwhist. Split-file layout introduced by issue
// #4366; card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** A card played into the current German Whist trick. */
export interface GermanWhistTrickCard {
  playerIdx: number;
  card: Card;
}

/** One seat at a German Whist table. */
export interface GermanWhistPlayer {
  id: number;
  isHuman: boolean;
  /** Hand size. The only hand information exposed for the CPU. */
  cardCount: number;
  /** Populated for the human player only; empty for the CPU. */
  cards: Card[];
  /** Total tricks taken across both halves. */
  trickCount: number;
  /** Tricks taken in the SECOND half only — the ones that decide the game. */
  scoringTricks: number;
}

/** A suggested card to play. */
export interface GermanWhistHint {
  cardIndex?: number;
  /**
   * Why that card. The aim inverts between the halves: `germanWhistTakeUpCard`
   * (win it, the face-up card is worth having), `germanWhistDuck` (lose on
   * purpose, it isn't), `germanWhistWinTrick` (second half — every trick scores).
   */
  reason: string;
}

/** Full German Whist game state returned from the API. */
export interface GermanWhistResponse extends BaseGameResponse {
  players: GermanWhistPlayer[];
  /** `0` = Draw (first 13 tricks, no score), `1` = Scoring, `2` = GameEnd. */
  phase: number;
  trickNumber: number;
  currentPlayerIdx: number;
  leadPlayerIdx: number;
  currentTrick: GermanWhistTrickCard[];
  /** Suit of the first face-up card; trump for the whole hand. */
  trumpSuit: number;
  /** The card the first half is played for. Absent once the stock is exhausted. */
  upCard?: Card;
  stockCount: number;
  /** Hand indices you may legally play. Following suit is compulsory. */
  validPlays: number[];
  gameEndFlag: boolean;
  /** `-1` until decided. */
  winnerIdx: number;
  hint?: GermanWhistHint;
}
