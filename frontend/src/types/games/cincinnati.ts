// Type declarations for cincinnati. Split-file layout introduced by issue #4366;
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Phases, matching the Go domain. */

/** Betting actions, matching the Go domain (shared betting constants). */

/** Hole cards per seat. **Five, not Hold'em's two.** */

/** A suggestion for the current decision. */
export interface CincinnatiHint {
  /** `fold`, `check`, `call`, `bet` or `raise`. */
  action: string;
  reason: string;
}

/** One seat at the table. */
export interface CincinnatiSeat {
  name: string;
  isHuman: boolean;
  chips: number;
  /** Net stack change since the ante was posted at hand start. */
  netChange: number;
  bet: number;
  /**
   * The five hole cards.
   *
   * **Empty for CPU seats until showdown.** The server withholds them rather
   * than relying on the page to hide them — with five cards each, shipping
   * them would settle the hand.
   */
  cards: Card[];
  folded: boolean;
  allIn: boolean;
  isTurn: boolean;
  /** Set at showdown only. */
  handRank: number;
  /** The best five of the ten cards. Set at showdown only. */
  bestHand: Card[];
  wonAmount: number;
}

/** Cincinnati game settings. */
export interface CincinnatiConfig {
  seats: number;
  initialChips: number;
  ante: number;
}

/** Response payload for `/cincinnati/exec`. */
export interface CincinnatiResponse extends BaseGameResponse {
  /** 0=Deal, 1=Betting, 2=Showdown, 3=GameEnd. */
  phase: number;
  seats: CincinnatiSeat[];
  /** Only the community cards already turned face up. */
  community: Card[];
  /** How many are face up (0..5). */
  revealedCount: number;
  /** How many there will be in total (5) — lets the page show the rounds left. */
  communityTotal: number;
  pot: number;
  currentBet: number;
  /** What the human seat must put in to call. */
  toCall: number;
  raiseCount: number;
  /** False once the per-round raise cap is reached. **Do not re-derive.** */
  canRaise: boolean;
  turnSeat: number;
  humanSeat: number;
  isHumanTurn: boolean;
  handNumber: number;
  remainingCards: number;
  winnerSeat: number;
  gameEndFlag: boolean;
  config?: CincinnatiConfig;
}
