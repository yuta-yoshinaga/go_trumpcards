// Type declarations for tablanet. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Tablanet game phase (0=Play 1=GameEnd). */
export type TablanetPhaseValue = 0 | 1;

/** A Tablanet player's public/own state. Cards are non-empty only for the human. */
export interface TablanetPlayer {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  /** Number of cards captured so far this game. */
  capturedCount: number;
  /** Number of Tabla sweeps (clearing the table with a single non-Jack card). */
  tablaCount: number;
  /** Final score (populated at game end). */
  score: number;
}

/** Per-game scoring breakdown for Tablanet (surfaced at game end). */
export interface TablanetScoreDetail {
  /** Captured card counts per seat, keyed by seat index. */
  cards: Record<number, number>;
  /** Ace counts per seat, keyed by seat index. */
  aces: Record<number, number>;
  /** Jack counts per seat, keyed by seat index. */
  jacks: Record<number, number>;
  /** Tabla sweep counts per seat, keyed by seat index. */
  tablas: Record<number, number>;
  /** Seat index holding the 10♦, or -1. */
  hasTenDiamonds: number;
  /** Seat index holding the 2♣, or -1. */
  hasTwoClubs: number;
  /** Seat index with the (unique) most captured cards, or -1 on a tie. */
  mostCards: number;
  /** Points gained per seat this game, keyed by seat index. */
  gained: Record<number, number>;
}

/** Tablanet game configuration. */
export interface TablanetConfig {
  cpuDifficulty: number;
}

/** A suggested hint for Tablanet, computed by the backend. */
export interface TablanetHint {
  cardIndices: number[];
  /** Suggested table-card indices to capture (present for a capture hint). */
  tableIndices?: number[];
  /** i18n reason suffix identifier. */
  reason: string;
}

/**
 * Full Tablanet (Tablić) game state returned from the API.
 *
 * Tablanet is a 4-player (1 human + 3 CPU, individual scoring) fishing/capture
 * game on a standard 52-card deck. Each player is dealt four cards with four
 * face-up on the table. A number card captures same-rank cards and any table
 * subset summing to its value; a Jack sweeps the whole table (except other Jacks).
 * Clearing the table with a single non-Jack card scores a "Tabla" bonus. When the
 * stock is exhausted the game ends (`gameEndFlag` true) and scores are tallied.
 */
export interface TablanetResponse extends BaseGameResponse {
  players: TablanetPlayer[];
  phase: TablanetPhaseValue;
  /** Number of packs dealt so far (deal counter). */
  roundNumber: number;
  /** Seat index of the player whose turn it is to act. */
  currentTurn: number;
  /** Cards currently face-up on the table. */
  tableCards: Card[];
  /** Seat index of the last player who captured, or -1. */
  lastCaptureIdx: number;
  /** Number of cards left in the stock. */
  remainingDeck: number;
  /** Indices in the human's hand that are legal to play (non-empty on human turn). */
  playableIndices: number[];
  /** Map of hand index -> table indices that hand card can capture (human turn). */
  captureOptions: Record<number, number[]>;
  /** Seat indices of the winner(s) at game end. */
  winners: number[];
  /** Whether the game has ended (stock exhausted and scored). */
  gameEndFlag: boolean;
  lastDealDetail?: TablanetScoreDetail | null;
  /** Whether it is currently the human's turn to act. */
  isHumanTurn: boolean;
  hint?: TablanetHint | null;
  config: TablanetConfig;
}

// --- Solo Whist ---
