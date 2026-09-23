// Type declarations for tapptarock. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** One card played into the current trick, with the seat that played it. */
export interface TappTarockTrickCard {
  playerIdx: number;
  card: Card;
}

/** TappTarock phase value (0=Bid 1=Talon 2=Play 3=TrickEnd 4=RoundEnd 5=GameEnd). */
export type TappTarockPhaseValue = 0 | 1 | 2 | 3 | 4 | 5;

/** A TappTarock seat. Hand `cards` are non-empty only for the human. */
export interface TappTarockPlayer {
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
export interface TappTarockBreakdown {
  contract: number;
  /** Declarer side's card points (for Trischaken, the loser's). */
  teamPoints: number;
  /** Points needed to succeed (more than this). */
  threshold: number;
  won: boolean;
  solo: boolean;
  base: number;
  /** Per-seat change, in seat order. Always sums to zero. */
  seats: number[];
  /** Trischaken's biggest taker (-1 for other contracts). */
  loser: number;
  /** Contract identifier ("trischaken" | "dreier" | "solo" | "pass"). */
  name: string;
}

/** A suggested hint for TappTarock, computed by the backend. */
export interface TappTarockHint {
  bid?: number | null;
  cardIndex?: number | null;
  discardIndices: number[];
  /** i18n reason suffix identifier. */
  reason: string;
}

/** TappTarock game configuration. */
export interface TappTarockConfig {
  cpuDifficulty: number;
  targetDeals: number;
}

/**
 * Full TappTarock (タップ・タロック) game state returned from the API.
 *
 * An Austrian three-player tarock game on the 54-card tarock pack. The highest
 * bidder plays against the other two players; everyone passing creates the
 * Trischaken house contract.
 */
export interface TappTarockResponse extends BaseGameResponse {
  players: TappTarockPlayer[];
  phase: TappTarockPhaseValue;
  roundNumber: number;
  totalRounds: number;
  trickNumber: number;
  currentPlayerIdx: number;
  dealerIdx: number;
  bidPlayerIdx: number;
  highestBid: number;
  /** Declarer seat, or -1 while undecided / under Trischaken. */
  declarerIdx: number;
  contract: number;
  /** Contract identifier ("pass" | "trischaken" | "dreier" | "solo"). */
  contractName: string;
  talonCount: number;
  currentTrick: TappTarockTrickCard[];
  lastTrickWinner: number;
  lastTrickCards: Card[];
  outcome: number;
  breakdown?: TappTarockBreakdown | null;
  /** Hand indices the human may legally play right now. */
  playableIndices: number[];
  /** Hand indices the human declarer may legally bury during the Talon phase. */
  discardableIndices: number[];
  gameEndFlag: boolean;
  winnerPlayer: number;
  isHumanTurn: boolean;
  hint?: TappTarockHint | null;
  config: TappTarockConfig;
}
