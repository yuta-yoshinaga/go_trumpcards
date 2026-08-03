// Type declarations for vira. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Vira game phase (0=Bid 1=Play 2=TrickEnd 3=RoundEnd 4=GameEnd). */
export type ViraPhaseValue = 0 | 1 | 2 | 3 | 4;

/** A Vira player's public/own state. Cards are non-empty only for the human. */
export interface ViraPlayer {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  trickCount: number;
  /** Cumulative match score of this individual player. */
  score: number;
  /** Whether this player is the round's declarer (plays alone vs the 2 defenders). */
  isDeclarer: boolean;
}

/** A card played into the current Vira trick. */
export interface ViraTrickCard {
  playerIdx: number;
  card: Card;
}

/** Vira game configuration. */
export interface ViraConfig {
  cpuDifficulty: number;
  targetRounds: number;
}

/** A suggested hint for Vira, computed by the backend. */
export interface ViraHint {
  cardIndices: number[];
  /** i18n reason suffix identifier. */
  reason: string;
}

/** Full Vira game state returned from the API. */
export interface ViraResponse extends BaseGameResponse {
  players: ViraPlayer[];
  phase: ViraPhaseValue;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  leadPlayerIdx: number;
  dealerIdx: number;
  /** Seat index of the round's declarer, or -1 before bidding resolves. */
  declarerIdx: number;
  /** Winning contract (0=Pass 1=Gask 2=Solo 3=Misère 4=Vira). */
  contract: number;
  /** Trump suit (0=none during bid / Misère, else 1=♠ 2=♣ 3=♥ 4=♦). */
  trumpSuit: number;
  /** Each player's bid this round (0-4) — [p0, p1, p2]. */
  bids: number[];
  /**
   * The running pot.
   *
   * **It carries forward between rounds**, including through an all-pass round,
   * so the same contract is worth more after a run of failures. Showing it is
   * what makes that difference legible.
   */
  pot: number;
  /** Change in each player's score from the round just settled — [p0, p1, p2]. */
  lastRoundDelta: number[];
  /** Whether the contract just settled was made. */
  lastRoundMade: boolean;
  currentTrick: ViraTrickCard[];
  /** Cumulative match scores per player — [p0, p1, p2]. */
  playerScores: number[];
  /** Tricks captured per player this round — [p0, p1, p2]. */
  roundTricks: number[];
  /** Indices in the human's hand that are legal to play (non-empty on human Play turn). */
  playableIndices: number[];
  gameEndFlag: boolean;
  /** Winning player seat index, or -1 until the game ends. */
  winnerPlayer: number;
  /** Whether it is currently the human's turn to play a card. */
  isHumanTurn: boolean;
  /** Whether it is currently the human's turn to bid. */
  isHumanBidTurn: boolean;
  hint?: ViraHint | null;
  config: ViraConfig;
}
