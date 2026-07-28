// Type declarations for durak. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Durak player data. */
export interface DurakPlayerData {
  id: number;
  isHuman: boolean;
  isFinished: boolean;
  cardCount: number;
  cards: Card[];
}

/** Durak table pair (attack + optional defense card). */
export interface DurakTablePair {
  attack: Card;
  defense: Card | null;
}

/** Durak CPU/human action record. */
export interface DurakAction {
  playerIdx: number;
  actionType: number; // 0=attack, 1=defend, 2=pass, 3=take
  card: Card | null;
  attackIdx: number;
}

/** Durak game rule configuration. */
export interface DurakConfig {
  playerCount: number;
  cpuDifficulty: number;
  transferEnabled: boolean;
}

/** Input type alias for Durak configuration. */
export type DurakConfigInput = DurakConfig;

/** Full Durak game state returned from the API. */
export interface DurakResponse extends BaseGameResponse {
  players: DurakPlayerData[];
  currentTurn: number;
  phase: number;
  attackerIdx: number;
  defenderIdx: number;
  tablePairs: DurakTablePair[];
  trumpSuit: string;
  trumpCard: Card | null;
  stockCount: number;
  loserIdx: number;
  gameEndFlag: boolean;
  config: DurakConfig;
  cpuActions: DurakAction[];
  humanAction: DurakAction | null;
  boutNumber: number;
  sortMode: number;
}

// --- Forty Thieves (フォーティシーブス) ---
