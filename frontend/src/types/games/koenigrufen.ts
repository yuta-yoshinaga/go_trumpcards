// Type declarations for koenigrufen. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Königrufen game phase (0=Bid 1=Call-a-king 2=Talon/discard 3=Play 4=TrickEnd 5=RoundEnd 6=GameEnd). */
export type KoenigrufenPhaseValue = 0 | 1 | 2 | 3 | 4 | 5 | 6;

/** A Königrufen player's public/own state. Cards are non-empty only for the human. */
export interface KoenigrufenPlayer {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  trickCount: number;
  /** Card-points captured in tricks so far this deal. */
  cardPoints: number;
  /** Cumulative match score of this individual player. */
  score: number;
  /** Whether this player is the declarer (contract holder) this deal. */
  isDeclarer: boolean;
  /** Whether this player is the declarer's secret partner. Only ever true once partnerRevealed is true. */
  isPartner: boolean;
}

/** A card played into the current Königrufen trick. */
export interface KoenigrufenTrickCard {
  playerIdx: number;
  card: Card;
}

/** Königrufen game configuration. */
export interface KoenigrufenConfig {
  cpuDifficulty: number;
  /** Number of deals that make up the match; the highest cumulative score wins. */
  targetDeals: number;
}

/** A suggested hint for Königrufen, computed by the backend. */
export interface KoenigrufenHint {
  /** Suggested bid value during the Bid phase, or null/undefined outside it. */
  bid?: number | null;
  /** Suggested King suit to call during the Call phase (1-4), or null/undefined outside it. */
  callSuit?: number | null;
  cardIndices: number[];
  /** i18n reason suffix identifier. */
  reason: string;
}

/**
 * Full Königrufen (ケーニッヒルーフェン) game state returned from the API.
 *
 * Königrufen is a 4-player tarock trick-taker on the 54-card tarock deck. After
 * the auction the declarer calls a King (King-calling / Rufer); whoever holds
 * that King becomes the declarer's secret partner. The declarer then exchanges
 * the talon (buries 6 cards) and the four play out the tricks. The partner's
 * identity stays hidden (`partnerIdx` is -1) until `partnerRevealed` is true.
 */
export interface KoenigrufenResponse extends BaseGameResponse {
  players: KoenigrufenPlayer[];
  phase: KoenigrufenPhaseValue;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  leadPlayerIdx: number;
  dealerIdx: number;
  /** Seat index of the player currently to bid (Bid phase). */
  bidPlayerIdx: number;
  /** The highest bid so far (0=none/pass, 1=Rufer). */
  highestBid: number;
  /** Seat index of the current highest bidder, or -1. */
  highestBidder: number;
  /** Seat index of the declarer, or -1 until decided. */
  declarerIdx: number;
  /** The declared contract (0=None, 1=Rufer). */
  contract: number;
  /** The called King's suit (1=Spade 2=Clover 3=Heart 4=Diamond), or -1 until called. */
  calledKing: number;
  /** Seat index of the declarer's secret partner — always -1 until partnerRevealed is true. */
  partnerIdx: number;
  /** Whether the secret partner has been revealed (partner shown only when true). */
  partnerRevealed: boolean;
  /** Number of cards in the talon (buried stash). */
  talonCount: number;
  /** The talon cards — non-empty only when revealed to a human declarer during the discard. */
  talon: Card[];
  /** Seat index that receives the talon's stashed card points (declarer or -1). */
  stashOwner: number;
  currentTrick: KoenigrufenTrickCard[];
  /** Cumulative match score per player — [p0, p1, p2, p3]. */
  playerScores: number[];
  /** Seat index of the last trick winner, or -1. */
  lastTrickWinner: number;
  /** Deal outcome (0=None, 1=Win/contract made, 2=Loss/contract failed). */
  outcome: number;
  /** Match result from the human's perspective (-1 lose, 0 none, 1 win). */
  result: number;
  /** Indices in the human's hand that are legal to play (non-empty on human Play turn). */
  playableIndices: number[];
  gameEndFlag: boolean;
  /** Winning player seat index, or -1 until the game ends (also -1 on a draw). */
  winnerPlayer: number;
  /** Whether it is currently the human's turn to act (play). */
  isHumanTurn: boolean;
  /** Whether it is currently the human's turn to bid. */
  isHumanBidTurn: boolean;
  /** Whether it is the human declarer's turn to call a King (Call phase). */
  isHumanCall: boolean;
  /** Whether it is the human declarer's turn to discard the talon (6 cards). */
  isHumanDiscard: boolean;
  hint?: KoenigrufenHint | null;
  config: KoenigrufenConfig;
}

// --- Cego (Baden Tarock) ---
