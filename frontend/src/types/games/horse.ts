// Type declarations for horse. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** H.O.R.S.E. phase value (0=Hand 1=HandEnd 2=GameEnd). */
export type HorsePhaseValue = 0 | 1 | 2;

/**
 * One seat at the H.O.R.S.E. table.
 *
 * `cards` holds only what that seat reveals: every card for the human, and in
 * the stud disciplines the face-up door cards for everyone else. It is empty
 * for a CPU seat in the hold'em-style disciplines, where nothing is face up.
 */
export interface HorseSeat {
  id: number;
  name: string;
  isHuman: boolean;
  chips: number;
  cards: Card[];
}

/** The five disciplines, in the order they rotate. */
export interface HorseConfigResponse {
  seats: number;
  initialChips: number;
  handsPerDiscipline: number;
}

/**
 * Response from the /horse/exec endpoint.
 *
 * The betting rules themselves belong to whichever discipline is running, so
 * this response carries the orchestrator's own state — which discipline, which
 * hand, the seats and their chips — plus the cards the table reveals.
 */
export interface HorseResponse extends BaseGameResponse {
  seats: HorseSeat[];
  phase: HorsePhaseValue;
  /** Discipline index (0=Hold'em 1=Omaha Hi-Lo 2=Razz 3=Stud 4=Stud Hi-Lo). */
  discipline: number;
  /** The letter of H.O.R.S.E. currently being played. */
  disciplineLetter: string;
  /** Stable key for the current discipline ("holdem", "omahaHiLo", ...). */
  disciplineName: string;
  handInDiscipline: number;
  handNumber: number;
  currentTurn: number;
  humanSeat: number;
  isHumanTurn: boolean;
  /** Shared cards; always empty in the stud disciplines. */
  communityCards: Card[];
  pot: number;
  /** Chips the human must put in to call; 0 means checking is legal. */
  toCall: number;
  /** Smallest raise the running discipline accepts. */
  minRaise: number;
  /** Phase number reported by the running discipline. */
  tablePhase: number;
  gameEndFlag: boolean;
  /** Seat with the most chips once the match is over, -1 while it runs. */
  winnerSeat: number;
  config: HorseConfigResponse;
}
