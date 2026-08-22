// Type declarations for put. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Put player data with per-hand baza (trick) count. */
export interface PutPlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  trickCount: number;
}

/** A card played in a Put baza (trick). */
export interface PutTrickCard {
  playerIdx: number;
  card: Card;
}

/** Put game configuration. */
export interface PutConfig {
  cpuDifficulty: number;
  matchTarget: number;
}

/** A suggested hint for Put (action is play / call / accept / decline). */
export interface PutHint {
  action: string;
  cardIndex?: number;
  reason: string;
}

/** Full Put game state returned from the API. */
export interface PutResponse extends BaseGameResponse {
  players: PutPlayerData[];
  phase: number;
  handNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  /** Player who must respond to a pending Put call; -1 when not awaiting a response. */
  responderIdx: number;
  currentTrick: PutTrickCard[];
  /** Outcome of each completed baza this hand: 0/1 = winner, -1 = parda (tie). */
  trickResults: number[];
  leadPlayerIdx: number;
  manoIdx: number;
  dealerIdx: number;
  /** Current agreed stake for the hand (1..4). */
  handStake: number;
  /** Accepted betting level (0=none, 1=Put, 2=Reput, 3=Vale Cuatro). */
  acceptedLevel: number;
  /** Proposed level while awaiting a response (0 otherwise). */
  pendingLevel: number;
  /** Index of the player whose Put call is pending (-1 otherwise). */
  putCallerIdx: number;
  /** Whether the human may declare / raise Put right now. */
  canDeclarePut: boolean;
  /** Points needed to win the match. */
  matchTarget: number;
  /** Cumulative match points [p0, p1]. */
  matchPoints: number[];
  /** Winner of the most recent hand (-1 = unresolved). */
  handWinnerIdx: number;
  gameEndFlag: boolean;
  /** -1 = unfinished. */
  winnerIdx: number;
  config: PutConfig;
  hint?: PutHint;
}

// --- Poker Squares (ポーカー・スクエア) ---
