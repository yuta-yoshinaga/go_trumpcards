// Type declarations for bolivia. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Bolivia game configuration. */
export interface BoliviaConfig {
  cpuDifficulty: number;
  pointLimit: number;
}

/** Meld kinds, matching the Go domain. **Three, not two.** */
export const BOLIVIA_MELD_KIND = {
  SET: 0,
  /** A suited run with no wilds. Seven of them is an **escalera**. */
  ESCALERA: 1,
  /** Wilds only. Seven of them is a **bolivia** — the game's namesake. */
  WILD: 2,
} as const;

/**
 * A single meld on the table.
 *
 * **`isEscalera` and `isBolivia` are different melds**, and the distinction
 * matters: going out requires an escalera (a wild-free suited seven), while a
 * bolivia (seven wilds) is merely the heaviest score. Conflating them is the
 * mistake the linked issue makes.
 */
export interface BoliviaMeldData {
  cards: Card[];
  /** 0 = set, 1 = escalera, 2 = wild-only. */
  kind: number;
  isNatural: boolean;
  /** Seven same-rank cards. */
  isCanasta: boolean;
  /** Seven suited cards in sequence, no wilds. **Going out needs one.** */
  isEscalera: boolean;
  /** Seven wild cards. Worth 2500 and closed to further cards. */
  isBolivia: boolean;
  rank: number;
}

/** Bolivia player data with melds, red 3s, and partnership (team) affiliation. */
export interface BoliviaPlayerData {
  id: number;
  team: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  melds: BoliviaMeldData[];
  red3Count: number;
  red3s: Card[];
  roundScore: number;
  cumulativeScore: number;
  hasCanasta: boolean;
  /** Holds a completed escalera. **This is what going out requires.** */
  hasEscalera: boolean;
  /** Holds a completed bolivia (seven wilds). Heaviest score, not a go-out key. */
  hasBolivia: boolean;
  hasInitMeld: boolean;
}

/** Full Bolivia game state returned from the API. */
export interface BoliviaResponse extends BaseGameResponse {
  players: BoliviaPlayerData[];
  teamScores: number[];
  phase: number;
  roundNumber: number;
  currentPlayerIdx: number;
  discardTop: Card | null;
  drawPileCount: number;
  discardPileCount: number;
  isFrozen: boolean;
  gameEndFlag: boolean;
  winnerIdx: number;
  config: BoliviaConfig;
}

// --- Hand and Foot (ハンド・アンド・フット) ---
