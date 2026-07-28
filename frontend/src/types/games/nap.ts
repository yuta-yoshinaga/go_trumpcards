// Type declarations for nap. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Nap game phase (0=Bid 1=Play 2=TrickEnd 3=RoundEnd 4=GameEnd). */
export type NapPhaseValue = 0 | 1 | 2 | 3 | 4;

/** A Nap player's public/own state. Cards are non-empty only for the human. */
export interface NapPlayer {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  trickCount: number;
  /** Cumulative chip score of this individual player. */
  score: number;
  /** Whether this player is the round's declarer (plays to make the bid). */
  isDeclarer: boolean;
}

/** A card played into the current Nap trick. */
export interface NapTrickCard {
  playerIdx: number;
  card: Card;
}

/** Nap game configuration. */
export interface NapConfig {
  cpuDifficulty: number;
  targetPoints: number;
}

/** A suggested hint for Nap, computed by the backend. */
export interface NapHint {
  cardIndices: number[];
  /** i18n reason suffix identifier. */
  reason: string;
}

/** Full Nap game state returned from the API. */
export interface NapResponse extends BaseGameResponse {
  players: NapPlayer[];
  phase: NapPhaseValue;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  leadPlayerIdx: number;
  dealerIdx: number;
  /** Seat index of the round's declarer, or -1 before bidding resolves. */
  declarerIdx: number;
  /** Winning contract (0=Pass 2=Two 3=Three 4=Four 5=Nap; the value is the bid trick count). */
  contract: number;
  /** Trump suit (0 during bid, else 1=♠ 2=♣ 3=♥ 4=♦ in play). */
  trumpSuit: number;
  /** Each player's bid this round (0/2/3/4/5) — [p0, p1, p2, p3]. */
  bids: number[];
  currentTrick: NapTrickCard[];
  /** Cumulative chip scores per player — [p0, p1, p2, p3]. */
  playerScores: number[];
  /** Tricks captured per player this round — [p0, p1, p2, p3]. */
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
  hint?: NapHint | null;
  config: NapConfig;
}
