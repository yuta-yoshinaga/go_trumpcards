// Type declarations for skitgubbe. Split-file layout introduced by issue
// #4366; card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** One Skitgubbe seat. */
export interface SkitgubbePlayer {
  id: number;
  isHuman: boolean;
  /** Hand size. Always sent, including while {@link SkitgubbePlayer.hidden} is true. */
  cardCount: number;
  /** Empty while {@link SkitgubbePlayer.hidden} is true. */
  cards: Card[];
  /**
   * Cards taken in phase one. Public, and worth reading: they become this
   * seat's phase-two hand, so collecting too much is a liability.
   */
  collectedCount: number;
  /** Out of cards in phase two. */
  finished: boolean;
  /** Whether this seat's HAND is withheld. */
  hidden: boolean;
}

/** Suggested move for the human seat. */
export interface SkitgubbeHintPayload {
  cardIndex?: number;
  /** Whether picking the pile up is the suggestion. */
  pickUp: boolean;
  /** Reason identifier, e.g. `skitgubbe.hint.beat`. */
  reason: string;
}

/** Full Skitgubbe game state returned from the API. */
export interface SkitgubbeResponse extends BaseGameResponse {
  players: SkitgubbePlayer[];
  /** 0 = Collect, 1 = Shed, 2 = GameEnd. */
  phase: number;
  currentPlayerIdx: number;
  stockCount: number;
  /**
   * Fixed by the LAST card drawn from the stock, so it is -1 for most of phase
   * one.
   */
  trumpSuit: number;
  /**
   * Phase-one cards on the table. Grows in pairs while stunsa (equal ranks)
   * keeps bouncing.
   */
  duel: Card[];
  duelLeader: number;
  /** Phase-two cards on the table. The last one is what must be beaten. */
  pile: Card[];
  /**
   * Hand indices that may be played. Carries the beat rule (higher of the same
   * suit, or a trump), so the page never re-derives it.
   */
  validIndices: number[];
  /**
   * True only in phase two, with a non-empty pile and no playable card.
   * Ducking is never lawful, so this is exactly the negation of
   * {@link SkitgubbeResponse.validIndices} being non-empty.
   */
  canPickUp: boolean;
  gameEndFlag: boolean;
  /** The Skitgubbe — the last seat holding cards; -1 while undecided. */
  loserIdx: number;
  hint?: SkitgubbeHintPayload;
}
