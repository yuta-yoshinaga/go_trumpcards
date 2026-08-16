// Type declarations for niuniu. Split-file layout introduced by issue
// #4366; card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** One Niu Niu hand. */
export interface NiuNiuHand {
  /**
   * Null entries while {@link NiuNiuHand.hidden} is true. The server does not
   * send the cards of a hand the player may not see; only the COUNT survives.
   */
  cards: (Card | null)[];
  bet: number;
  /**
   * Positions of the three cards that summed to a multiple of ten. Empty for a
   * no-bull, and empty while hidden. Sent so the UI can highlight them without
   * redoing the combination search.
   */
  comboIdx: number[];
  /** 0 = no bull, 1-9 = niu 1 through niu 9, 10 = niu niu. 0 while hidden. */
  rank: number;
  /**
   * Locale-independent rank key: `"none"`, `"niuniu"`, or `"n1"`..`"n9"`. Empty
   * while hidden. Render it with `niuniuRankText` -- the server used to send the
   * Japanese label itself, which ignored the locale (#5567).
   */
  rankKey: string;
  /** Payout multiplier for winning with this rank. 0 while hidden. */
  multiplier: number;
  payout: number;
  /** While true the hand's cards, rank and combo are withheld by the server. */
  hidden: boolean;
}

/** One Niu Niu seat. */
export interface NiuNiuSeat {
  name: string;
  isCpu: boolean;
  /** Absent before the deal, and for the banker's seat. */
  hand?: NiuNiuHand;
}

/** Full Niu Niu game state returned from the API. */
export interface NiuNiuResponse extends BaseGameResponse {
  seats: NiuNiuSeat[];
  bankerHand?: NiuNiuHand;
  bankerIdx: number;
  chips: number;
  /**
   * Largest payout multiplier in the table. A stake is only legal if the stack
   * covers `stake * maxMultiplier`, because a banker's Niu Niu takes three
   * times the stake -- so the cap on what may be bet is chips / this, not
   * chips. Sent so the figure is not written down twice.
   */
  maxMultiplier: number;
  /**
   * The banker's rank key once the round settles, empty before that. Replaces
   * the pre-built Japanese summary the server used to send (#5567).
   */
  bankerRankKey: string;
  phase: number;
}
