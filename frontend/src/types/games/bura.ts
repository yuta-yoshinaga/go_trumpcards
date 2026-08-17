// Type declarations for bura. Split-file layout introduced by issue
// #4366; card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** One Bura seat. */
export interface BuraPlayer {
  id: number;
  isHuman: boolean;
  /**
   * How many cards the seat holds. Always sent, including while
   * {@link BuraPlayer.hidden} is true -- how many cards someone holds is
   * public, and the UI needs it to draw the right number of card backs.
   */
  cardCount: number;
  /**
   * Empty while {@link BuraPlayer.hidden} is true. The server does not send
   * the cards of a hand the player may not see.
   */
  cards: Card[];
  points: number;
  /** While true the seat's cards are withheld by the server. */
  hidden: boolean;
}

/** Suggested play for the human seat. */
export interface BuraHintPayload {
  /** Hand indices to play. Empty when the suggestion is to claim or declare. */
  cardIndices?: number[];
  /** Reason identifier, e.g. `bura.hint.lead`. */
  reason: string;
}

/** Full Bura game state returned from the API. */
export interface BuraResponse extends BaseGameResponse {
  players: BuraPlayer[];
  phase: number;
  trickNumber: number;
  currentPlayerIdx: number;
  leadPlayerIdx: number;
  /** Cards led this trick. Empty between tricks. */
  currentLead: Card[];
  trumpSuit: number;
  /** Absent once the face-up indicator has been drawn into a hand. */
  trumpCard?: Card;
  stockRemaining: number;
  /**
   * Card points needed to claim a win. Sent rather than hardcoded so the
   * figure is not written down in two places.
   */
  winThreshold: number;
  gameEndFlag: boolean;
  winnerIdx: number;
  /** True when the stock emptied with nobody claiming: nobody wins. */
  isDraw: boolean;
  /**
   * Identifiers of the combinations that win outright, in the order the server
   * checks them. Sent so the explanation cannot list a different set from the
   * one the rules actually recognise.
   */
  winningCombinations: string[];
  hint?: BuraHintPayload;
}
