// Type declarations for manille. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Manille game phase (0=Play 1=TrickEnd 2=RoundEnd 3=GameEnd). */
export type ManillePhaseValue = 0 | 1 | 2 | 3;

/** A Manille player's public/own state. Cards are non-empty only for the human. */
export interface ManillePlayer {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  trickCount: number;
  /** Cumulative match score of the team this player belongs to. */
  teamScore: number;
}

/** A card played into the current Manille trick. */
export interface ManilleTrickCard {
  playerIdx: number;
  card: Card;
}

/** Manille game configuration. */
export interface ManilleConfig {
  cpuDifficulty: number;
  targetPoints: number;
}

/** A suggested hint for Manille, computed by the backend. */
export interface ManilleHint {
  cardIndices: number[];
  /** i18n reason suffix identifier. */
  reason: string;
}

/** Full Manille game state returned from the API. */
export interface ManilleResponse extends BaseGameResponse {
  players: ManillePlayer[];
  phase: ManillePhaseValue;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  leadPlayerIdx: number;
  dealerIdx: number;
  /** Trump suit (1=♠ 2=♣ 3=♥ 4=♦). */
  trumpSuit: number;
  currentTrick: ManilleTrickCard[];
  /** Cumulative match scores per team — [team0, team1]. */
  teamScores: number[];
  /** Card points captured per team this round — [team0, team1]. */
  roundCardPoints: number[];
  /** Indices in the human's hand that are legal to play (non-empty on human Play turn). */
  playableIndices: number[];
  gameEndFlag: boolean;
  /** Winning team index, or -1 until the game ends. */
  winnerTeam: number;
  /** Whether it is currently the human's turn to act. */
  isHumanTurn: boolean;
  hint?: ManilleHint | null;
  config: ManilleConfig;
}

// --- Sedma ---
