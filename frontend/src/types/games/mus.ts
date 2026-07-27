// Type declarations for mus. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/**
 * Mus game phase
 * (0=Mus 1=Discard 2=Grande 3=Chica 4=Pares 5=Juego 6=Showdown 7=RoundEnd 8=GameEnd).
 */
export type MusPhaseValue = 0 | 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8;

/**
 * A Mus player's public/own state. `cards` is populated for the human at all
 * times and for opponents only once the phase reaches Showdown (>=6).
 */
export interface MusPlayer {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  /** The score (amarrakos) of the team this player belongs to. */
  teamScore: number;
}

/**
 * Result of one of the four betting rounds (Grande / Chica / Pares / Juego).
 * `kind` identifies the round, `stake` the amarrakos awarded, `team` the winner.
 */
export interface MusRoundResult {
  kind: number;
  stake: number;
  team: number;
}

/** Mus game configuration. */
export interface MusConfig {
  cpuDifficulty: number;
  targetAmarrakos: number;
}

/** A suggested hint for Mus, computed by the backend. */
export interface MusHint {
  /** Whether the hint recommends calling Mus / exchanging (Mus phase). */
  mus: boolean;
  /** Suggested bet action (0=paso 1=envido 2=ordago 3=quiero 4=noquiero). */
  action: number;
  /** Suggested Envido amount. */
  amount: number;
  /** Suggested card indices to discard (Discard phase). */
  indices: number[];
  /** i18n reason suffix identifier. */
  reason: string;
}

/** Full Mus game state returned from the API. */
export interface MusResponse extends BaseGameResponse {
  players: MusPlayer[];
  phase: MusPhaseValue;
  roundNumber: number;
  /** Index of the mano (lead) player. */
  manoIdx: number;
  /** Team that currently holds the active bet, or -1 when none. */
  betTeam: number;
  /** Pending stake amount (-1=ordago/all-in, 0=none). */
  pendingStake: number;
  /** Team that placed the most recent bet, or -1. */
  lastBettorTeam: number;
  /** Index of the player to act in the Mus phase. */
  musTurn: number;
  /** Index of the player to act in the Discard phase. */
  discardTurn: number;
  /** Number of Mus/exchange cycles completed this round. */
  musCycle: number;
  /** Team amarrakos (scores) — [team0, team1]. */
  amarrakos: number[];
  /** Per-round results indexed by Grande/Chica/Pares/Juego. */
  results: MusRoundResult[];
  gameEndFlag: boolean;
  /** Winning team index, or -1 until the game ends. */
  winnerTeam: number;
  /** Team the human player belongs to, or -1 when none. */
  humanTeam: number;
  /** Whether it is currently the human's turn to act. */
  isHumanTurn: boolean;
  /** Whether the Paso bet action is legal for the human right now. */
  canPaso: boolean;
  /** Whether the Envido bet action is legal for the human right now. */
  canEnvido: boolean;
  /** Whether the Ordago (all-in) bet action is legal for the human right now. */
  canOrdago: boolean;
  /** Whether the Quiero (accept) bet action is legal for the human right now. */
  canQuiero: boolean;
  /** Whether the No Quiero (decline) bet action is legal for the human right now. */
  canNoQuiero: boolean;
  hint?: MusHint | null;
  config: MusConfig;
}

// --- Doppelkopf ---
