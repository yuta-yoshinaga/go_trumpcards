// Type declarations for doppelkopf. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Doppelkopf game phase (0=Play 1=TrickEnd 2=RoundEnd 3=GameEnd). */
export type DoppelkopfPhaseValue = 0 | 1 | 2 | 3;

/** A Doppelkopf player's public/own state. Cards are non-empty only for the human. */
export interface DoppelkopfPlayer {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  trickCount: number;
  chips: number;
  /** Whether this player is on the Re team. False until teams are revealed. */
  isRe: boolean;
}

/** A card played into the current Doppelkopf trick. */
export interface DoppelkopfTrickCard {
  playerIdx: number;
  card: Card;
}

/** Doppelkopf game configuration. */
export interface DoppelkopfConfig {
  cpuDifficulty: number;
  baseChips: number;
  startChips: number;
  targetChips: number;
}

/** A suggested hint for Doppelkopf, computed by the backend. */
export interface DoppelkopfHint {
  cardIndices: number[];
  reason: string;
}

/** Full Doppelkopf game state returned from the API. */
export interface DoppelkopfResponse extends BaseGameResponse {
  players: DoppelkopfPlayer[];
  phase: DoppelkopfPhaseValue;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  leadPlayerIdx: number;
  dealerIdx: number;
  currentTrick: DoppelkopfTrickCard[];
  /** Each player's Re-team membership; all false until teams are revealed (4 elements). */
  reTeam: boolean[];
  /** Whether one player holds both ♣Q (a solo Re). */
  soloRe: boolean;
  /** Whether the Re/Kontra teams have been revealed. */
  teamsRevealed: boolean;
  /** Whether Re has been announced this round. */
  reAnnounced: boolean;
  /** Whether Kontra has been announced this round. */
  kontraAnnounced: boolean;
  /** Whether the human may announce Re/Kontra right now (first trick only). */
  canAnnounce: boolean;
  /** Whether the human is on the Re team. Always known, even before reveal. */
  youAreRe: boolean;
  /** Indices in the human's hand that are legal to play (non-empty on human Play turn). */
  playableIndices: number[];
  /** Card points captured by the Re team this round. */
  roundRePoints: number;
  /** Whether the Re team won the round. */
  roundReWon: boolean;
  /** Game points awarded for this round. */
  roundGamePoints: number;
  gameEndFlag: boolean;
  /** Winning player index, or -1 until the game ends. */
  winnerIdx: number;
  hint?: DoppelkopfHint | null;
  config: DoppelkopfConfig;
}

// --- Tute ---
