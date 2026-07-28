// Type declarations for truco. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Truco player data with per-hand baza (trick) count. */
export interface TrucoPlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  trickCount: number;
}

/** A card played in a Truco baza (trick). */
export interface TrucoTrickCard {
  playerIdx: number;
  card: Card;
}

/** Truco game configuration. */
export interface TrucoConfig {
  cpuDifficulty: number;
  matchTarget: number;
}

/** A suggested hint for Truco (action is play / call / accept / decline). */
export interface TrucoHint {
  action: string;
  cardIndex?: number;
  reason: string;
}

/** Full Truco game state returned from the API. */
export interface TrucoResponse extends BaseGameResponse {
  players: TrucoPlayerData[];
  phase: number;
  handNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  /** Player who must respond to a pending Truco call; -1 when not awaiting a response. */
  responderIdx: number;
  currentTrick: TrucoTrickCard[];
  /** Outcome of each completed baza this hand: 0/1 = winner, -1 = parda (tie). */
  trickResults: number[];
  leadPlayerIdx: number;
  manoIdx: number;
  dealerIdx: number;
  /** Current agreed stake for the hand (1..4). */
  handStake: number;
  /** Accepted betting level (0=none, 1=Truco, 2=Retruco, 3=Vale Cuatro). */
  acceptedLevel: number;
  /** Proposed level while awaiting a response (0 otherwise). */
  pendingLevel: number;
  /** Index of the player whose Truco call is pending (-1 otherwise). */
  trucoCallerIdx: number;
  /** Whether the human may declare / raise Truco right now. */
  canDeclareTruco: boolean;
  /** Points needed to win the match. */
  matchTarget: number;
  /** Cumulative match points [p0, p1]. */
  matchPoints: number[];
  /** Winner of the most recent hand (-1 = unresolved). */
  handWinnerIdx: number;
  gameEndFlag: boolean;
  /** -1 = unfinished. */
  winnerIdx: number;
  config: TrucoConfig;
  hint?: TrucoHint;
}

// --- Poker Squares (ポーカー・スクエア) ---
