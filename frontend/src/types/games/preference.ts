// Type declarations for preference. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Préférence game phase (0=Bid 1=Play 2=TrickEnd 3=RoundEnd 4=GameEnd). */
export type PreferencePhaseValue = 0 | 1 | 2 | 3 | 4;

/** A Préférence player's public/own state. Cards are non-empty only for the human. */
export interface PreferencePlayer {
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

/** A card played into the current Préférence trick. */
export interface PreferenceTrickCard {
  playerIdx: number;
  card: Card;
}

/** Préférence game configuration. */
export interface PreferenceConfig {
  cpuDifficulty: number;
  targetPoints: number;
}

/** A suggested hint for Préférence, computed by the backend. */
export interface PreferenceHint {
  cardIndices: number[];
  /** i18n reason suffix identifier. */
  reason: string;
}

/** Full Préférence game state returned from the API. */
export interface PreferenceResponse extends BaseGameResponse {
  players: PreferencePlayer[];
  phase: PreferencePhaseValue;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  leadPlayerIdx: number;
  dealerIdx: number;
  /** Seat index of the round's declarer, or -1 before bidding resolves. */
  declarerIdx: number;
  /** Winning contract (0=Pass 1=Six 2=Misère 3=Seven 4=Eight). */
  contract: number;
  /** Trump suit (0=none during bid / Misère, else 1=♠ 2=♣ 3=♥ 4=♦). */
  trumpSuit: number;
  /** Each player's bid this round (0-4) — [p0, p1, p2]. */
  bids: number[];
  currentTrick: PreferenceTrickCard[];
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
  hint?: PreferenceHint | null;
  config: PreferenceConfig;
}

// --- Nap (Napoleon) ---
