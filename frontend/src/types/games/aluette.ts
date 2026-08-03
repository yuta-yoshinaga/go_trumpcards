// Type declarations for aluette. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card, CardDesign } from '../common';

/** Aluette game phase (0=Play 1=TrickEnd 2=RoundEnd 3=GameEnd). */
export type AluettePhaseValue = 0 | 1 | 2 | 3;

/** An Aluette player's public/own state. Cards are non-empty only for the human. */
export interface AluettePlayer {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  trickCount: number;
  /** Team number: opposite seats partner, so seats 0/2 are one team and 1/3 the other. */
  team: number;
  isDealer: boolean;
}

/** A card played into the current Aluette trick. */
export interface AluetteTrickCard {
  playerIdx: number;
  card: Card;
}

/** Aluette game configuration. */
export interface AluetteConfig {
  cpuDifficulty: number;
  /** Menes the match is played to. */
  targetPoints: number;
}

/**
 * One of the six named cards that outrank the whole deck.
 *
 * The backend sends the table on every response rather than the frontend
 * keeping its own copy — the six cards and their order *are* the game, so a
 * second copy would eventually disagree with the domain.
 */
export interface AluetteLuette {
  /** Suit in the same wire form as a played card, e.g. "DIAMOND". */
  design: CardDesign;
  value: number;
  /** Traditional name: Monsieur, Madame, Borgne, Vache, GrandNeuf, PetitNeuf. */
  name: string;
}

/** A suggested hint for Aluette, computed by the backend. */
export interface AluetteHint {
  cardIndices: number[];
  /** i18n reason suffix identifier. */
  reason: string;
}

/** Full Aluette game state returned from the API. */
export interface AluetteResponse extends BaseGameResponse {
  players: AluettePlayer[];
  phase: AluettePhaseValue;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  leadPlayerIdx: number;
  dealerIdx: number;
  currentTrick: AluetteTrickCard[];
  /** Cumulative match score per team — [team0, team1]. */
  teamScores: number[];
  /** Tricks captured per seat this mene — [p0, p1, p2, p3]. */
  roundTricks: number[];
  /** Seat that took the previous trick, or -1 before the first is resolved. */
  lastTrickWinner: number;
  /**
   * Indices in the human's hand that are legal to play. Aluette has no follow
   * obligation, so on a human turn this is the whole hand.
   */
  playableIndices: number[];
  /** The six luettes, strongest first. Mirrors `domain.AluetteLuetteTable`. */
  luettes: AluetteLuette[];
  gameEndFlag: boolean;
  /** Winning team, or -1 while undecided and also on a draw. */
  winnerTeam: number;
  isHumanTurn: boolean;
  hint?: AluetteHint | null;
  config: AluetteConfig;
}

/** Number of players. Opposite seats partner into two fixed teams. */
export const ALUETTE_PLAYER_COUNT = 4;

/** Cards dealt to each player. 5 tricks per mene. */
export const ALUETTE_HAND_SIZE = 5;

/** The team a seat belongs to. Mirrors `domain.AluetteTeamOf`. */
export function aluetteTeamOf(seat: number): number {
  return seat % 2;
}

/**
 * The luette name for a card, or `undefined` when it is an ordinary card.
 *
 * Takes the table from the response rather than embedding a copy — see
 * {@link AluetteLuette}.
 */
export function aluetteLuetteName(luettes: AluetteLuette[], card: Card): string | undefined {
  return luettes.find((l) => l.design === card.design && l.value === card.value)?.name;
}
