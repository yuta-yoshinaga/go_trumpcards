// Type declarations for carioca. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Carioca meld: a set (trío) or run (escala) of cards laid down on the table. */
export interface CariocaMeld {
  cards: Card[];
}

/** Carioca contract slot: a single requirement of the round's contract. */
export interface CariocaContractSlot {
  /** 0 = set (same rank), 1 = run (same suit consecutive). */
  kind: number;
  /** Number of cards required to fill this slot. */
  size: number;
}

/** Carioca player state. */
export interface CariocaPlayer {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  melds: CariocaMeld[];
  /** Whether the player has met this round's contract. */
  contractMet: boolean;
  roundScore: number;
  cumulativeScore: number;
}

/** Carioca game configuration. */
export interface CariocaConfig {
  playerCount: number;
  cpuDifficulty: number;
  failContractPenalty: number;
}

/** Carioca API response. */
export interface CariocaResponse extends BaseGameResponse {
  players: CariocaPlayer[];
  /** 0 = draw, 1 = play, 2 = round end, 3 = game end. */
  phase: number;
  roundNumber: number;
  totalRounds: number;
  currentPlayerIdx: number;
  discardTop: Card | null;
  drawPileCount: number;
  gameEndFlag: boolean;
  winnerIdx: number;
  roundWinnerIdx: number;
  /** The current round's contract (sequence of slots to satisfy). */
  contractSlots: CariocaContractSlot[];
  config: CariocaConfig;
}

// --- Kalooki (カルーキ) ---
