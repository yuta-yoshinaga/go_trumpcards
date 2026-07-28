// Type declarations for klaverjas. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Klaverjas game phase (0=Play 1=TrickEnd 2=RoundEnd 3=GameEnd). */
export type KlaverjasPhaseValue = 0 | 1 | 2 | 3;

/** A Klaverjas player's public/own state. Cards are non-empty only for the human. */
export interface KlaverjasPlayer {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  trickCount: number;
  /** Cumulative match score of the team this player belongs to. */
  teamScore: number;
}

/** A card played into the current Klaverjas trick. */
export interface KlaverjasTrickCard {
  playerIdx: number;
  card: Card;
}

/** Klaverjas game configuration. */
export interface KlaverjasConfig {
  cpuDifficulty: number;
  targetPoints: number;
}

/** A suggested hint for Klaverjas, computed by the backend. */
export interface KlaverjasHint {
  cardIndices: number[];
  /** i18n reason suffix identifier. */
  reason: string;
}

/** Full Klaverjas game state returned from the API. */
export interface KlaverjasResponse extends BaseGameResponse {
  players: KlaverjasPlayer[];
  phase: KlaverjasPhaseValue;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  leadPlayerIdx: number;
  dealerIdx: number;
  /** Trump suit (1=♠ 2=♣ 3=♥ 4=♦). */
  trumpSuit: number;
  currentTrick: KlaverjasTrickCard[];
  /** Cumulative match scores per team — [team0, team1]. */
  teamScores: number[];
  /** Card points captured per team this round — [team0, team1]. */
  roundCardPoints: number[];
  /** Roem (run/marriage) bonus points per team this round — [team0, team1]. */
  roundRoem: number[];
  /** Indices in the human's hand that are legal to play (non-empty on human Play turn). */
  playableIndices: number[];
  gameEndFlag: boolean;
  /** Winning team index, or -1 until the game ends. */
  winnerTeam: number;
  /** Whether it is currently the human's turn to act. */
  isHumanTurn: boolean;
  hint?: KlaverjasHint | null;
  config: KlaverjasConfig;
}

// --- Manille ---
