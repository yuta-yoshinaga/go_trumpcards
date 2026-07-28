// Type declarations for king. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/**
 * King game phase, mirrored from the Go domain string constants
 * (sync: internal/domain/King.go).
 *   - `selectContract` — the dealer chooses the deal's contract.
 *   - `play` — the 13-trick must-follow play phase.
 *   - `dealEnd` — one deal finished; scores settled, waiting for the next deal.
 *   - `gameEnd` — all seven contracts played; the match is over.
 */
export type KingPhaseValue = 'selectContract' | 'play' | 'dealEnd' | 'gameEnd';

/** A King player's public/own state. Cards are non-empty only for the human. */
export interface KingPlayer {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  /** Tricks captured this deal. */
  trickCount: number;
  /** Cumulative match score (King is trick-avoidance, so higher = fewer penalties). */
  totalScore: number;
}

/** A card played into the current King trick. */
export interface KingTrickCard {
  playerIdx: number;
  card: Card;
}

/** King game configuration. */
export interface KingConfig {
  cpuDifficulty: number;
}

/** Per-deal scoring detail surfaced at the end of each deal. */
export interface KingDealDetail {
  /** Contract index played this deal (0=No Tricks … 6=King Trump). */
  contract: number;
  /** Trump suit for the King (Trump) contract (1=♠ 2=♣ 3=♥ 4=♦), else -1. */
  trumpSuit: number;
  /** Seat index of the dealer who chose the contract. */
  dealerIdx: number;
  /** Points gained per player this deal, keyed by seat index. */
  gained: Record<number, number>;
}

/** A suggested hint for King, computed by the backend. */
export interface KingHint {
  cardIndices: number[];
  /** i18n reason suffix identifier (e.g. `avoid_low`, `win_high`). */
  reason: string;
}

/**
 * Full King game state returned from the API.
 *
 * King is a 4-player 52-card compendium trick-avoidance game. Each match is
 * exactly seven deals; the dealer of each deal chooses one of seven unused
 * contracts (0=No Tricks … 6=King Trump), then all four seats play thirteen
 * must-follow tricks. The highest total score (i.e. the fewest penalty points)
 * wins the match.
 */
export interface KingResponse extends BaseGameResponse {
  players: KingPlayer[];
  phase: KingPhaseValue;
  /** Current deal index (0-based) within the seven-deal match. */
  dealNumber: number;
  /** Total deals per match (always 7). */
  totalDeals: number;
  dealerIdx: number;
  /** Seat index whose turn it currently is. */
  currentTurn: number;
  /** Contract chosen this deal (0..6), or -1 before selection. */
  currentContract: number;
  /** Trump suit for the King (Trump) contract (1=♠ 2=♣ 3=♥ 4=♦), else -1. */
  trumpSuit: number;
  trickNumber: number;
  currentTrick: KingTrickCard[];
  lastTrick: KingTrickCard[];
  /** Seat index of the last trick winner, or -1. */
  lastTrickWinner: number;
  /** Which of the seven contracts have already been played this match. */
  usedContracts: boolean[];
  /** Indices in the human's hand that are legal to play (non-empty on human Play turn). */
  playableIndices: number[];
  gameEndFlag: boolean;
  config: KingConfig;
  /** Seat indices of the match winner(s); empty until the game ends. */
  roundWinners: number[];
  /** Scoring detail for the most recently completed deal, or null. */
  lastDealDetail?: KingDealDetail | null;
  /** Whether it is currently the human's turn to act. */
  isHumanTurn: boolean;
  hint?: KingHint | null;
}

// --- Tysiąc (Thousand) ---
