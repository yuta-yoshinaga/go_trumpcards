// Type declarations for rikken. Split-file layout introduced by issue #4366;
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Trump value meaning "no trump" — the Misere contracts. Suits are 1..4. */
export const RIKKEN_NO_TRUMP = -1;

/**
 * The contract ladder, weakest first. **The number is the bidding strength**,
 * and bids may only rise.
 *
 * Note that contracts of opposite kinds share the ladder: Rik and Solo are
 * about *taking* tricks, Misere and Open Misere about taking none.
 */
export const RIKKEN_CONTRACTS = [
  { contract: 1, key: 'contract.rik' },
  { contract: 2, key: 'contract.misere' },
  { contract: 3, key: 'contract.solo' },
  { contract: 4, key: 'contract.openMisere' },
] as const;

/** One seat at a Rikken table. */
export interface RikkenPlayer {
  id: number;
  isHuman: boolean;
  cardCount: number;
  /** Only the human seat carries its cards. */
  cards: Card[];
  trickCount: number;
  /** Running score. **Negative is normal** — scoring is zero-sum. */
  score: number;
  /** Sides are set by the contract, not by seat. */
  isDeclarerSide: boolean;
  hasPassed: boolean;
}

/** A card played into the current trick. */
export interface RikkenTrickCard {
  playerIdx: number;
  card: Card;
}

/** A suggestion: either a contract to bid or a card to play. */
export interface RikkenHint {
  contract?: number;
  cardIndex?: number;
  /** `rikkenBidStrength` or `rikkenFollowSuit`. */
  reason: string;
}

/** Rikken game settings. */
export interface RikkenConfig {
  rounds: number;
}

/** Response payload for `/rikken/exec`. */
export interface RikkenResponse extends BaseGameResponse {
  players: RikkenPlayer[];
  /** 0=Bid, 1=Call, 2=Play, 3=RoundEnd, 4=GameEnd. */
  phase: number;
  validPlays: number[];
  dealerIdx: number;
  /** 0=none, 1=Rik, 2=Misere, 3=Solo, 4=Open Misere. */
  contract: number;
  /** -1 until the auction resolves. */
  declarerIdx: number;
  /** -1 until the called card is played. */
  partnerIdx: number;
  /** The ace called to find a partner. Rik only. */
  calledCard?: Card;
  /** 1..4, or -1 for no trump. */
  trumpSuit: number;
  currentTurn: number;
  isHumanTurn: boolean;
  currentTrick: RikkenTrickCard[];
  lastTrick: RikkenTrickCard[];
  lastTrickWinner: number;
  trickCount: number;
  declarerTricks: number;
  roundNumber: number;
  gameEndFlag: boolean;
  winnerIdx: number;
  config?: RikkenConfig;
}
