// Type declarations for zwanzigerrufen. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** One card played into the current trick, with the seat that played it. */
export interface ZwanzigerrufenTrickCard {
  playerIdx: number;
  card: Card;
}

/** Zwanzigerrufen phase value (0=Bid 1=Talon 2=Play 3=TrickEnd 4=RoundEnd 5=GameEnd). */
export type ZwanzigerrufenPhaseValue = 0 | 1 | 2 | 3 | 4 | 5;

/** A Zwanzigerrufen seat. Hand `cards` are non-empty only for the human. */
export interface ZwanzigerrufenPlayer {
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
  /** True only once the partner has been revealed. */
  isPartner: boolean;
}

/** Per-deal settlement breakdown. */
export interface ZwanzigerrufenBreakdown {
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
  /** Contract identifier ("trischaken" | "rufer" | "solo" | "pass"). */
  name: string;
}

/** A suggested hint for Zwanzigerrufen, computed by the backend. */
export interface ZwanzigerrufenHint {
  bid?: number | null;
  cardIndex?: number | null;
  discardIndices: number[];
  /** i18n reason suffix identifier. */
  reason: string;
}

/** Zwanzigerrufen game configuration. */
export interface ZwanzigerrufenConfig {
  cpuDifficulty: number;
  targetDeals: number;
}

/**
 * Full Zwanzigerrufen (ツヴァンツィガールーフェン) game state returned from the API.
 *
 * An Austrian calling tarock for four on the 54-card tarock pack. Two things
 * separate it from Königrufen: **the declarer calls the trump XX** (stepping
 * down to XIX/XVIII when they hold it themselves), and **when everyone passes
 * the deal becomes Trischaken**, where each player plays for themselves and
 * whoever takes the most card points loses.
 *
 * `partnerIdx` is `-1` until `partnerRevealed` — the secret partner is never
 * put on the wire before the called trump is played.
 */
export interface ZwanzigerrufenResponse extends BaseGameResponse {
  players: ZwanzigerrufenPlayer[];
  phase: ZwanzigerrufenPhaseValue;
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
  /** Contract identifier ("pass" | "trischaken" | "rufer" | "solo"). */
  contractName: string;
  /** Called trump number (18-20), or -1 when nothing was called. */
  calledTrump: number;
  /** Partner seat — always -1 until `partnerRevealed`. */
  partnerIdx: number;
  partnerRevealed: boolean;
  talonCount: number;
  currentTrick: ZwanzigerrufenTrickCard[];
  lastTrickWinner: number;
  lastTrickCards: Card[];
  outcome: number;
  breakdown?: ZwanzigerrufenBreakdown | null;
  /** Hand indices the human may legally play right now. */
  playableIndices: number[];
  gameEndFlag: boolean;
  winnerPlayer: number;
  isHumanTurn: boolean;
  hint?: ZwanzigerrufenHint | null;
  config: ZwanzigerrufenConfig;
}
