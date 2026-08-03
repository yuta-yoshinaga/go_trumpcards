// Type declarations for chineseten. Split-file layout introduced by issue
// #4366; card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** One card, with the value the server already worked out. */
export interface ChineseTenCard extends Card {
  /** What this card is worth. Non-zero only for red cards. */
  points: number;
  /** Whether the card is a scoring (heart or diamond) card. */
  isRed: boolean;
}

/** One Chinese Ten seat. */
export interface ChineseTenPlayer {
  id: number;
  isHuman: boolean;
  /** Hand size. Always sent, including while {@link ChineseTenPlayer.hidden} is true. */
  cardCount: number;
  /** Empty while {@link ChineseTenPlayer.hidden} is true. */
  cards: ChineseTenCard[];
  /**
   * ALWAYS present for both seats. Which cards have already gone is public,
   * and reading it is how a fishing game is played.
   */
  captured: ChineseTenCard[];
  score: number;
  /** Whether this seat's HAND is withheld. Captures are never hidden. */
  hidden: boolean;
}

/** Suggested move for the human seat. */
export interface ChineseTenHintPayload {
  cardIndex?: number;
  layoutIndex?: number;
  /** Reason identifier, e.g. `chineseten.hint.play`. */
  reason: string;
}

/** Full Chinese Ten game state returned from the API. */
export interface ChineseTenResponse extends BaseGameResponse {
  players: ChineseTenPlayer[];
  layout: ChineseTenCard[];
  /** 0 = Play, 1 = Select, 2 = GameEnd. */
  phase: number;
  currentPlayerIdx: number;
  stockCount: number;
  /** The card awaiting a capture choice, if any. */
  pendingCard?: ChineseTenCard;
  /**
   * Layout indices the pending card may take. Carries BOTH capture rules
   * (sum-to-ten for A-9, same rank for 10-K), so the page never re-derives a
   * pair of rules that do not overlap.
   */
  selectableIndices: number[];
  /** The score that draws the game — half the red cards' 210. */
  tieScore: number;
  gameEndFlag: boolean;
  winnerIdx: number;
  hint?: ChineseTenHintPayload;
}
