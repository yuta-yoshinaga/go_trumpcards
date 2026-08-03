// Type declarations for tarocchini. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Tarocchini game phase (0=Scarto 1=Play 2=TrickEnd 3=RoundEnd 4=GameEnd). */
export type TarocchiniPhaseValue = 0 | 1 | 2 | 3 | 4;

/** A Tarocchini player's public/own state. Cards are non-empty only for the human. */
export interface TarocchiniPlayer {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  trickCount: number;
  /** Team number: opposite seats partner, so seats 0/2 are one team and 1/3 the other. */
  team: number;
  isDealer: boolean;
}

/** A card played into the current Tarocchini trick. */
export interface TarocchiniTrickCard {
  playerIdx: number;
  card: Card;
}

/** Tarocchini game configuration. */
export interface TarocchiniConfig {
  cpuDifficulty: number;
  /** Constrained to a multiple of the player count so the deal goes round once. */
  targetRounds: number;
}

/** A suggested hint for Tarocchini, computed by the backend. */
export interface TarocchiniHint {
  cardIndices: number[];
  /** i18n reason suffix identifier. */
  reason: string;
}

/** Full Tarocchini game state returned from the API. */
export interface TarocchiniResponse extends BaseGameResponse {
  players: TarocchiniPlayer[];
  phase: TarocchiniPhaseValue;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  leadPlayerIdx: number;
  dealerIdx: number;
  /** Number of cards the dealer buried this round (0 until the scarto is done). */
  scartoCount: number;
  currentTrick: TarocchiniTrickCard[];
  /** Cumulative match score per team — [team0, team1]. */
  teamScores: number[];
  /** Tricks captured per seat this round — [p0, p1, p2, p3]. */
  roundTricks: number[];
  /** Seat that took the previous trick, or -1 before the first is resolved. */
  lastTrickWinner: number;
  /** Indices in the human's hand that are legal to play (non-empty on a human Play turn). */
  playableIndices: number[];
  gameEndFlag: boolean;
  /** Winning team, or -1 while undecided and also on a draw. */
  winnerTeam: number;
  isHumanTurn: boolean;
  /** Whether it is the human's turn to bury the surplus cards. */
  isHumanScarto: boolean;
  hint?: TarocchiniHint | null;
  config: TarocchiniConfig;
}

/** Cards the dealer must bury. 62 does not divide by 4, so 2 are always left over. */
export const TAROCCHINI_SURPLUS = 2;

/** Number of players. Opposite seats partner into two fixed teams. */
export const TAROCCHINI_PLAYER_COUNT = 4;

/** The team a seat belongs to. Mirrors `domain.TarocchiniTeamOf`. */
export function tarocchiniTeamOf(seat: number): number {
  return seat % 2;
}
