// Type declarations for mushi. Split-file layout introduced by issue
// #4366; card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/**
 * One hanafuda card, with the identity the server already worked out.
 *
 * `category` and `points` ride along so the 40-card identity table exists in
 * exactly one place; the client never re-derives what a card is worth.
 */
export interface MushiCard extends Card {
  /**
   * The card's MONTH (1-12, with 6 and 7 absent). `design` cannot be used for
   * this: hanafuda encodes the month there, but the wire converts it to a suit
   * name, so the month is sent separately rather than decoded back out of
   * "CLOVER".
   */
  month: number;
  /** Index within the month, 1-4. */
  index: number;
  /** 0 = chaff, 1 = ribbon, 2 = animal, 3 = bright. */
  category: number;
  points: number;
  /** True only for November's lightning card, which takes any non-willow card. */
  isWild: boolean;
}

/** One Mushi seat. */
export interface MushiPlayer {
  id: number;
  isHuman: boolean;
  /** Hand size. Always sent, including while {@link MushiPlayer.hidden} is true. */
  cardCount: number;
  /** Empty while {@link MushiPlayer.hidden} is true. */
  cards: MushiCard[];
  /**
   * ALWAYS present for both seats. Captured cards are public in hanafuda --
   * reading what the opponent is collecting is how the game is played.
   */
  captured: MushiCard[];
  capturedPoints: number;
  /** Cumulative score across rounds. */
  score: number;
  /** The change from the round just settled. */
  roundResult: number;
  /** Whether this seat's HAND is withheld. Captured cards are never hidden. */
  hidden: boolean;
}

/** Suggested move for the human seat. */
export interface MushiHintPayload {
  cardIndex?: number;
  fieldIndex?: number;
  /** Reason identifier, e.g. `mushi.hint.play`. */
  reason: string;
}

/** Full Mushi game state returned from the API. */
export interface MushiResponse extends BaseGameResponse {
  players: MushiPlayer[];
  field: MushiCard[];
  /** 0 = Play, 1 = Select, 2 = WildSelect, 3 = RoundEnd, 4 = GameEnd. */
  phase: number;
  currentPlayerIdx: number;
  dealerIdx: number;
  roundNumber: number;
  targetRounds: number;
  stockCount: number;
  /** The card awaiting a capture choice, if any. */
  pendingCard?: MushiCard;
  /**
   * Field indices that may be taken right now. Carries the wild's "not another
   * willow" restriction, so the page never re-implements that rule.
   */
  selectableIndices: number[];
  gameEndFlag: boolean;
  winnerIdx: number;
  hint?: MushiHintPayload;
}
