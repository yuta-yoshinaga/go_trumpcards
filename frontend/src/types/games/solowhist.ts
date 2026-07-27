// Type declarations for solowhist. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Solo Whist game phase (0=Bid 1=Play 2=TrickEnd 3=RoundEnd 4=GameEnd). */
export type SoloWhistPhaseValue = 0 | 1 | 2 | 3 | 4;

/** A Solo Whist player's public/own state. Cards are non-empty only for the human. */
export interface SoloWhistPlayer {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  trickCount: number;
  /** Cumulative match score of this individual player. */
  score: number;
  /** Whether this player is the round's declarer (plays alone vs the 3 defenders). */
  isDeclarer: boolean;
}

/** A card played into the current Solo Whist trick. */
export interface SoloWhistTrickCard {
  playerIdx: number;
  card: Card;
}

/** Solo Whist game configuration. */
export interface SoloWhistConfig {
  cpuDifficulty: number;
  targetPoints: number;
}

/** A suggested hint for Solo Whist, computed by the backend. */
export interface SoloWhistHint {
  cardIndices: number[];
  /** i18n reason suffix identifier. */
  reason: string;
}

// --- Auction Forty-Fives ---

/** Full Solo Whist game state returned from the API. */
export interface SoloWhistResponse extends BaseGameResponse {
  players: SoloWhistPlayer[];
  phase: SoloWhistPhaseValue;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  leadPlayerIdx: number;
  dealerIdx: number;
  /** Seat index of the round's declarer, or -1 before bidding resolves. */
  declarerIdx: number;
  /** Winning contract (0=Pass 1=Solo 2=Misère 3=Abundance). */
  contract: number;
  /** Trump suit (0=none for Misère, else 1=♠ 2=♣ 3=♥ 4=♦). */
  trumpSuit: number;
  /** Each player's bid this round (0-3) — [p0, p1, p2, p3]. */
  bids: number[];
  currentTrick: SoloWhistTrickCard[];
  /** Cumulative match scores per player — [p0, p1, p2, p3]. */
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
  hint?: SoloWhistHint | null;
  config: SoloWhistConfig;
}

// --- Spades ---
