// Type declarations for troggu. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Troggu phase value (0=Bid 1=Play 2=TrickEnd 3=RoundEnd 4=GameEnd). */
export type TrogguPhaseValue = 0 | 1 | 2 | 3 | 4;

/** One card played into the current trick, with the seat that played it. */
export interface TrogguTrickCard {
  playerIdx: number;
  card: Card;
}

/** A Troggu seat. Hand `cards` are non-empty only for the human. */
export interface TrogguPlayer {
  id: number;
  isHuman: boolean;
  cardCount: number;
  /** Hand cards (populated only for the human). */
  cards: Card[];
  trickCount: number;
  /** Card points taken so far this deal. */
  cardPoints: number;
  /** Running match score. */
  score: number;
  isDeclarer: boolean;
}

/** Per-deal settlement breakdown. */
export interface TrogguBreakdown {
  contract: number;
  /** Contract identifier ("trois" | "solo" | "piccolo" | "misere" | "pass"). */
  contractName: string;
  declarerPoints: number;
  declarerTricks: number;
  /** Target value: card points for Solo, tricks for every other contract. */
  target: number;
  /** True when `target` counts tricks rather than points. */
  targetIsTricks: boolean;
  won: boolean;
  base: number;
  /** Per-seat change, in seat order. Always sums to zero. */
  seats: number[];
}

/** A suggested hint for Troggu, computed by the backend. */
export interface TrogguHint {
  bid?: number | null;
  cardIndex?: number | null;
  /** i18n reason suffix identifier. */
  reason: string;
}

/** Troggu game configuration. */
export interface TrogguConfig {
  cpuDifficulty: number;
  targetDeals: number;
}

/**
 * Full Troggu (トロッグ) game state returned from the API.
 *
 * A tarot-family trick-taking game from the Swiss Valais, played by four on the
 * 78-card tarot pack. **What separates it from other tarot games is that each
 * contract changes what winning means**: Solo takes most of the card points,
 * Trois needs three tricks, Piccolo needs *exactly one*, and Misère needs none
 * at all — so at the same table the goal can point either way. When everyone
 * passes the deal is thrown in and no score moves.
 */
export interface TrogguResponse extends BaseGameResponse {
  players: TrogguPlayer[];
  phase: TrogguPhaseValue;
  roundNumber: number;
  totalRounds: number;
  trickNumber: number;
  currentPlayerIdx: number;
  dealerIdx: number;
  bidPlayerIdx: number;
  highestBid: number;
  /** Declarer seat, or -1 while undecided or after a thrown-in deal. */
  declarerIdx: number;
  contract: number;
  contractName: string;
  talonCount: number;
  currentTrick: TrogguTrickCard[];
  lastTrickWinner: number;
  lastTrickCards: Card[];
  outcome: number;
  breakdown?: TrogguBreakdown | null;
  /** Hand indices the human may legally play right now. */
  playableIndices: number[];
  gameEndFlag: boolean;
  winnerPlayer: number;
  isHumanTurn: boolean;
  hint?: TrogguHint | null;
  config: TrogguConfig;
}
