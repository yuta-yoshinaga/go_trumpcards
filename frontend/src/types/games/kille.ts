// Type declarations for kille. Split-file layout introduced by issue #4366;
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** What happened in one exchange this round. */
export interface KilleEvent {
  /** `swap` | `satisfied` | `cuckoo` | `hussar` | `pig` | `pass` | `stock`. */
  kind: string;
  actor: number;
  /** Seat challenged; -1 for a swap with the stock. */
  target: number;
}

/** One Kille seat. Everyone holds exactly one card. */
export interface KillePlayer {
  id: number;
  isHuman: boolean;
  /**
   * The seat's single card, or null while it is face down. Carries
   * `glyph` / `label` / `color` / `deck` because the Kille pack is not the
   * French 52 (ADR-0033).
   */
  card: Card | null;
  /**
   * Effective rank, 0 while the card is face down. **Not the same as the card's
   * denomination**: a Harlequin received in an exchange scores 0 rather than 21.
   */
  strength: number;
  chips: number;
  /** Buy-backs used so far, out of three. */
  reentries: number;
  /**
   * What the next buy-back costs — one stake, then half the pot, then the whole
   * pot. 0 once all three are spent.
   */
  reentryCost: number;
  canReenter: boolean;
  isOut: boolean;
  /** `hussar` | `pig` | `''` (went out as the lowest). */
  knockedBy: string;
  isSatisfied: boolean;
  /** Out for good — eliminated and no buy-backs left. */
  isFinished: boolean;
  isCurrentTurn: boolean;
}

/** Full Kille game state returned from the API. */
export interface KilleResponse extends BaseGameResponse {
  players: KillePlayer[];
  /** 0 = Exchange, 1 = Showdown, 2 = GameEnd. */
  phase: number;
  roundNumber: number;
  currentPlayerIdx: number;
  /**
   * The dealer acts last and swaps with the **stock** rather than a neighbour,
   * so it is the one seat that cannot be challenged.
   */
  dealerIdx: number;
  stockCount: number;
  pot: number;
  /** What happened in this round's exchanges, in order. */
  events: KilleEvent[];
  /** Seats that went out this round — more than one when a Hussar or Pig fired. */
  loserIdxs: number[];
  gameEndFlag: boolean;
  /** Last player standing; -1 while the game is live. */
  winnerIdx: number;
}
