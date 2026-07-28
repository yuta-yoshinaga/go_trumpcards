// Type declarations for nertz. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Tableau card with face-up state in a Nertz player area. */
export interface NertzTableauCard {
  card: Card | null;
  faceUp: boolean;
}

/** Nertz player view (per-player tableau, nertz pile, waste, and stock). */
export interface NertzPlayerData {
  name: string;
  isHuman: boolean;
  deckIdx: number;
  score: number;
  nertzSize: number;
  nertzTop?: Card;
  tableau: NertzTableauCard[][];
  wasteTop?: Card;
  wasteSize: number;
  stockSize: number;
}

/** Nertz shared foundation pile. */
export interface NertzFoundationData {
  top?: Card;
  suit: number;
  size: number;
}

/** Nertz suggested move hint. */
export interface NertzHint {
  fromZone: string;
  fromCol: number;
  cardIndex: number;
  toZone: string;
  toCol: number;
}

/** Nertz local rule configuration. */
export interface NertzConfig {
  playerCount: number;
  drawCount: number;
  targetScore: number;
  cpuDifficulty: number;
  cpuTickMoves: number;
}

/** Source or target zone for a Nertz move. */
export interface NertzMoveZone {
  zone: 'nertz' | 'waste' | 'tableau' | 'foundation';
  col?: number;
  idx?: number;
  cardIndex?: number;
}

/** Full Nertz game state returned from the API. */
export interface NertzResponse extends BaseGameResponse {
  phase: number;
  roundNumber: number;
  winnerIdx: number;
  matchWinner: number;
  moveCount: number;
  canUndo: boolean;
  playerCount: number;
  drawCount: number;
  targetScore: number;
  cpuDifficulty: number;
  /** CPU per-tick budget (resolved from cpuDifficulty when 0). */
  cpuTickMoves: number;
  players: NertzPlayerData[];
  foundations: NertzFoundationData[];
  hint?: NertzHint;
}
