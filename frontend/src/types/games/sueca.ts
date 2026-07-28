// Type declarations for sueca. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Sueca game phase (0=Play 1=TrickEnd 2=RoundEnd 3=GameEnd). */
export type SuecaPhaseValue = 0 | 1 | 2 | 3;

/** A Sueca player's public/own state. Cards are non-empty only for the human. */
export interface SuecaPlayer {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  trickCount: number;
  /** Cumulative game points of the team this player belongs to. */
  teamGamePoints: number;
}

/** A card played into the current Sueca trick. */
export interface SuecaTrickCard {
  playerIdx: number;
  card: Card;
}

/** Sueca game configuration. */
export interface SuecaConfig {
  cpuDifficulty: number;
  targetGamePoints: number;
}

/** A suggested hint for Sueca, computed by the backend. */
export interface SuecaHint {
  cardIndices: number[];
  /** i18n reason suffix identifier. */
  reason: string;
}

/** Full Sueca game state returned from the API. */
export interface SuecaResponse extends BaseGameResponse {
  players: SuecaPlayer[];
  phase: SuecaPhaseValue;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  leadPlayerIdx: number;
  dealerIdx: number;
  /** Trump suit (1=♠ 2=♣ 3=♥ 4=♦). */
  trumpSuit: number;
  currentTrick: SuecaTrickCard[];
  /** Cumulative game points per team — [team0, team1]. */
  teamGamePoints: number[];
  /** Card points captured per team this round — [team0, team1]. */
  roundCardPoints: number[];
  /** Winning team of the most recent round, or -1 when undecided/draw. */
  roundWinnerTeam: number;
  /** Game points awarded for the most recent round. */
  roundGamePoints: number;
  /** Indices in the human's hand that are legal to play (non-empty on human Play turn). */
  playableIndices: number[];
  gameEndFlag: boolean;
  /** Winning team index, or -1 until the game ends. */
  winnerTeam: number;
  /** Whether it is currently the human's turn to act. */
  isHumanTurn: boolean;
  hint?: SuecaHint | null;
  config: SuecaConfig;
}

// --- Klaverjas ---
