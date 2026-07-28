// Type declarations for cego. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Cego game phase (0=Bid 1=Contract 2=Exchange 3=Play 4=TrickEnd 5=RoundEnd 6=GameEnd). */
export type CegoPhaseValue = 0 | 1 | 2 | 3 | 4 | 5 | 6;

/** A Cego player's public/own state. Cards are non-empty only for the human. */
export interface CegoPlayer {
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
}

/** A card played into the current Cego trick. */
export interface CegoTrickCard {
  playerIdx: number;
  card: Card;
}

/** Cego game configuration. */
export interface CegoConfig {
  cpuDifficulty: number;
  /** Number of deals that make up the match; the highest cumulative score wins. */
  targetDeals: number;
}

/** A suggested hint for Cego, computed by the backend. */
export interface CegoHint {
  /** Suggested bid value during the Bid phase, or null/undefined outside it. */
  bid?: number | null;
  /** Suggested contract during the Contract phase (1=Cego 2=Handspiel), or null/undefined outside it. */
  contract?: number | null;
  cardIndices: number[];
  /** i18n reason suffix identifier. */
  reason: string;
}

/**
 * Full Cego (チェゴ) game state returned from the API.
 *
 * Cego is a 4-player Baden tarock trick-taker on the 54-card tarock deck. One
 * declarer plays 1-vs-3. After the auction the declarer chooses a contract —
 * Cego (lay down all but one dealt card and pick up the 10-card blind) or
 * Handspiel (keep the dealt hand) — then the four play out 11 tricks. The
 * blind's contents are never revealed (only `blindCount`).
 */
export interface CegoResponse extends BaseGameResponse {
  players: CegoPlayer[];
  phase: CegoPhaseValue;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  leadPlayerIdx: number;
  dealerIdx: number;
  /** Seat index of the player currently to bid (Bid phase). */
  bidPlayerIdx: number;
  /** The highest bid so far (0=none/pass, 1=play). */
  highestBid: number;
  /** Seat index of the current highest bidder, or -1. */
  highestBidder: number;
  /** Seat index of the declarer, or -1 until decided. */
  declarerIdx: number;
  /** The winning bid (0=None, 1=Play). */
  contract: number;
  /** The chosen contract type (0=None, 1=Cego, 2=Handspiel). */
  contractType: number;
  /** Number of cards in the blind (Cego stash) — the contents stay hidden. */
  blindCount: number;
  /** The blind cards — always empty; the blind is hidden (use blindCount). */
  blind: Card[];
  /** Seat index that receives the blind's stashed card points (declarer side or opponents). */
  stashOwner: number;
  currentTrick: CegoTrickCard[];
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
  /** Whether it is the human declarer's turn to choose a contract (Contract phase). */
  isHumanContract: boolean;
  /** Whether it is the human declarer's turn to make the Cego exchange (keep exactly 1 card). */
  isHumanExchange: boolean;
  hint?: CegoHint | null;
  config: CegoConfig;
}

// --- Cinch ---
