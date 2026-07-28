// Type declarations for tute. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Tute game phase (0=Play 1=TrickEnd 2=RoundEnd 3=GameEnd). */
export type TutePhaseValue = 0 | 1 | 2 | 3;

/** A Tute player's public/own state. Cards are non-empty only for the human. */
export interface TutePlayer {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  trickCount: number;
  /** Cumulative score of the team this player belongs to. */
  teamScore: number;
}

/** A card played into the current Tute trick. */
export interface TuteTrickCard {
  playerIdx: number;
  card: Card;
}

/** Tute game configuration. */
export interface TuteConfig {
  cpuDifficulty: number;
  targetPoints: number;
}

/** A suggested hint for Tute, computed by the backend. */
export interface TuteHint {
  cardIndices: number[];
  /** Suggested marriage-declaration suit (0=none, 1=♠ 2=♣ 3=♥ 4=♦). */
  marriage: number;
  /** i18n reason suffix identifier. */
  reason: string;
}

/** Full Tute game state returned from the API. */
export interface TuteResponse extends BaseGameResponse {
  players: TutePlayer[];
  phase: TutePhaseValue;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  leadPlayerIdx: number;
  dealerIdx: number;
  /** Trump suit (1=♠ 2=♣ 3=♥ 4=♦). */
  trumpSuit: number;
  currentTrick: TuteTrickCard[];
  /** Declared-marriage suits; valid indices 1-4 (index 0 unused, 5 elements). */
  declaredSuits: boolean[];
  /** Team scores — [team0, team1]. */
  teamScores: number[];
  /** Card points captured per team this round — [team0, team1]. */
  roundTeamPoints: number[];
  /** Whether the human may declare a marriage (K+Q) right now. */
  canDeclareMarriage: boolean;
  /** Whether the human may declare Tute (four Kings or four Queens) for an instant win. */
  canDeclareTute: boolean;
  /** Indices in the human's hand that are legal to play (non-empty on human Play turn). */
  playableIndices: number[];
  gameEndFlag: boolean;
  /** Winning team index, or -1 until the game ends. */
  winnerTeam: number;
  /** Whether it is currently the human's turn to act. */
  isHumanTurn: boolean;
  hint?: TuteHint | null;
  config: TuteConfig;
}

// --- Sueca ---
