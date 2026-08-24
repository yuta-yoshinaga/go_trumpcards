// Type declarations for coinche. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Coinche player data with team, trick count, and hand. */
export interface CoinchePlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  team: number;
  trickCount: number;
}

/** A card played in a Coinche trick. */
export interface CoincheTrickCard {
  playerIdx: number;
  card: Card;
}

/** Coinche game configuration. */
export interface CoincheConfig {
  cpuDifficulty: number;
  targetScore: number;
  dixDeDer: number;
  enableBeloteRebelote: boolean;
}

/** The contract in play. Coinche settles on the bid, not on card points. */
export type CoincheDouble = 0 | 1 | 2;

/** A suggested hint for Coinche. */
export interface CoincheHint {
  cardIndex?: number;
  /** Suggested contract target; absent when passing is recommended. */
  bid?: number;
  /** Suggested trump suit, paired with {@link CoincheHint.bid}. */
  suit?: number;
  reason: string;
}

/** Full Coinche game state returned from the API. */
export interface CoincheResponse extends BaseGameResponse {
  players: CoinchePlayerData[];
  phase: number;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  bidPlayerIdx: number;
  dealerIdx: number;
  trumpSuit: number;
  /** The contract that won the auction; 0 until it settles. */
  contractPoints: number;
  /** Settlement multiplier: 1 normally, 2 after a coinche, 4 after a surcoinche. */
  multiplier: number;
  /** 0=none, 1=coinche, 2=surcoinche. */
  double: CoincheDouble;
  /**
   * Contracts this seat may still bid — those above the standing bid.
   * Empty once the auction has closed.
   */
  biddablePoints: number[];
  makerTeam: number;
  makerPlayerIdx: number;
  currentTrick: CoincheTrickCard[];
  teamScores: number[];
  roundPoints: number[];
  roundBeloteBonus: number[];
  gameEndFlag: boolean;
  winnerTeam: number;
  leadPlayerIdx: number;
  config: CoincheConfig;
  hint?: CoincheHint;
}

// --- Jass (Schieber) (ヤス) ---
