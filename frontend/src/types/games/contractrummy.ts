// Type declarations for contractrummy. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Contract Rummy meld: a set or run of cards laid down on the table. */
export interface ContractRummyMeld {
  cards: Card[];
}

/** Contract Rummy contract slot: a single requirement of the round's contract. */
export interface ContractRummyContractSlot {
  /** 0 = set (same rank), 1 = run (same suit consecutive). */
  kind: number;
  /** Number of cards required to fill this slot. */
  size: number;
}

/** Contract Rummy player state. */
export interface ContractRummyPlayer {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  melds: ContractRummyMeld[];
  /** Whether the player has met this round's contract. */
  contractMet: boolean;
  roundScore: number;
  cumulativeScore: number;
}

/** Contract Rummy game configuration. */
export interface ContractRummyConfig {
  cpuDifficulty: number;
  failContractPenalty: number;
}

/** Contract Rummy API response. */
export interface ContractRummyResponse extends BaseGameResponse {
  players: ContractRummyPlayer[];
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
  contractSlots: ContractRummyContractSlot[];
  config: ContractRummyConfig;
}

// --- Carioca (カリオカ) ---
