// Type declarations for cirulla. Follows the split-out convention of card.ts
// (issue #4366); card.ts re-exports this file.

import type { BaseGameResponse, Card } from '../common';

/** One seat at the two-handed table. */
export interface CirullaPlayer {
  id: number;
  isHuman: boolean;
  /** Hand cards. Populated only for the human. */
  cards: Card[];
  cardCount: number;
  /** Cards taken this round. */
  capturedCount: number;
  /** Denari (diamonds) taken this round. */
  denariCount: number;
  hasSetteBello: boolean;
  scope: number;
  bonusPoints: number;
  /** Running match score. */
  score: number;
  isDealer: boolean;
  /** Most recent deal bonus ("barsega" | "barsegon" | ""). */
  lastBonus: string;
}

/** One scoring category of a round. */
export interface CirullaScoreLine {
  /** "cards" | "denari" | "settebello" | "primiera" | "piccola" | "grande" | "scope" | "bonus". */
  key: string;
  /** Points per seat. */
  points: number[];
}

/** One round's scoring result. */
export interface CirullaResult {
  lines: CirullaScoreLine[];
  totals: number[];
  /** Seat that took all ten denari, or -1. This wins outright. */
  sweptDenari: number;
}

/** Cirulla game configuration. */
export interface CirullaConfig {
  cpuDifficulty: number;
  /** Points needed to win, 11-51. */
  targetScore: number;
}

/**
 * Full Cirulla game state returned from the API.
 *
 * The Ligurian cousin of Scopa. **The sum-to-fifteen capture is added to
 * Scopa's, not a replacement**, and an ace played to an ace-less table sweeps
 * everything — so the server enumerates the legal captures rather than leaving
 * the client to re-derive three interacting rules.
 */
export interface CirullaResponse extends BaseGameResponse {
  players: CirullaPlayer[];
  /** "play" | "roundEnd" | "gameEnd". */
  phase: string;
  roundNumber: number;
  dealerIdx: number;
  currentPlayerIdx: number;
  table: Card[];
  deckRemaining: number;
  /** Seat that captured last — the leftover table goes to them. */
  lastCapturer: number;
  /**
   * Legal capture groups for the human's cards, in hand order.
   *
   * `captureOptions[i]` is the list of table-index groups the i-th hand card
   * can take. Empty while it is not the human's turn.
   */
  captureOptions: number[][][];
  lastResult?: CirullaResult | null;
  gameEndFlag: boolean;
  winnerIdx: number;
  isHumanTurn: boolean;
  /** Hand index the backend suggests, or -1. */
  hintHandIdx: number;
  /** Table indices the backend suggests taking. */
  hintCaptureIdxs: number[];
  /** i18n reason identifier for the suggestion. */
  hintReason: string;
  config: CirullaConfig;
}
