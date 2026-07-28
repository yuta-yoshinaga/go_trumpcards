// Type declarations for marias. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Mariáš game phase (0=Play 1=TrickEnd 2=RoundEnd 3=GameEnd). */
export type MariasPhaseValue = 0 | 1 | 2 | 3;

/** A Mariáš player's public/own state. Cards are non-empty only for the human. */
export interface MariasPlayer {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  trickCount: number;
  /** Cumulative match (game-point) score of this individual player. */
  score: number;
  /** Whether this player is the round's Soloist (plays alone vs the 2 Defenders). */
  isSoloist: boolean;
}

/** A card played into the current Mariáš trick. */
export interface MariasTrickCard {
  playerIdx: number;
  card: Card;
}

/** Mariáš game configuration. */
export interface MariasConfig {
  cpuDifficulty: number;
  targetPoints: number;
}

/** A suggested hint for Mariáš, computed by the backend. */
export interface MariasHint {
  cardIndices: number[];
  /** i18n reason suffix identifier. */
  reason: string;
}

/** Full Mariáš game state returned from the API. */
export interface MariasResponse extends BaseGameResponse {
  players: MariasPlayer[];
  phase: MariasPhaseValue;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  leadPlayerIdx: number;
  dealerIdx: number;
  /** Seat index of the round's Soloist. */
  soloistIdx: number;
  /** Trump suit (1=♠ 2=♣ 3=♥ 4=♦). */
  trumpSuit: number;
  currentTrick: MariasTrickCard[];
  /** Cumulative match (game-point) scores per player — [p0, p1, p2]. */
  playerScores: number[];
  /** Card points captured per player this round — [p0, p1, p2]. */
  roundCardPoints: number[];
  /** Marriage (K+Q same suit) points scored per player this round — [p0, p1, p2]. */
  roundMarriage: number[];
  /** Seat index of the last (10th) trick winner, or -1. */
  lastTrickWinner: number;
  /** Indices in the human's hand that are legal to play (non-empty on human Play turn). */
  playableIndices: number[];
  gameEndFlag: boolean;
  /** Winning player seat index, or -1 until the game ends. */
  winnerPlayer: number;
  /** Whether it is currently the human's turn to act. */
  isHumanTurn: boolean;
  hint?: MariasHint | null;
  config: MariasConfig;
}

// --- King ---
