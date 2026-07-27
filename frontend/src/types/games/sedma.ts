// Type declarations for sedma. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Sedma game phase (0=Play 1=TrickEnd 2=RoundEnd 3=GameEnd). */
export type SedmaPhaseValue = 0 | 1 | 2 | 3;

/** A Sedma player's public/own state. Cards are non-empty only for the human. */
export interface SedmaPlayer {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  trickCount: number;
  /** Cumulative match score of the team this player belongs to. */
  teamScore: number;
}

/** A card played into the current Sedma trick. */
export interface SedmaTrickCard {
  playerIdx: number;
  card: Card;
}

/** Sedma game configuration. */
export interface SedmaConfig {
  cpuDifficulty: number;
  targetPoints: number;
}

/** A suggested hint for Sedma, computed by the backend. */
export interface SedmaHint {
  cardIndices: number[];
  /** i18n reason suffix identifier. */
  reason: string;
}

/**
 * Full Sedma game state returned from the API.
 *
 * Sedma is a Czech/Slovak no-trump capture trick-taker, so — unlike the
 * Manille shape it mirrors — there is intentionally no `trumpSuit` field.
 */
export interface SedmaResponse extends BaseGameResponse {
  players: SedmaPlayer[];
  phase: SedmaPhaseValue;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  leadPlayerIdx: number;
  dealerIdx: number;
  currentTrick: SedmaTrickCard[];
  /** Cumulative match scores per team — [team0, team1]. */
  teamScores: number[];
  /** Card points captured per team this round — [team0, team1]. */
  roundCardPoints: number[];
  /** Indices in the human's hand that are legal to play (every card is playable on the human Play turn). */
  playableIndices: number[];
  gameEndFlag: boolean;
  /** Winning team index, or -1 until the game ends. */
  winnerTeam: number;
  /** Whether it is currently the human's turn to act. */
  isHumanTurn: boolean;
  hint?: SedmaHint | null;
  config: SedmaConfig;
}

// --- Knockout Whist ---
