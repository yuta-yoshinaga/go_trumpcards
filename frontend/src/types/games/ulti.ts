// Type declarations for ulti. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Ulti game phase (0=Bid 1=Discard 2=Play 3=TrickEnd 4=RoundEnd 5=GameEnd). */
export type UltiPhaseValue = 0 | 1 | 2 | 3 | 4 | 5;

/** An Ulti player's public/own state. Cards are non-empty only for the human declarer. */
export interface UltiPlayer {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  trickCount: number;
  /** Card-points captured in tricks so far this deal (A/10 = 10 each, +10 last trick). */
  cardPoints: number;
  /** Cumulative coin balance across the match. */
  coins: number;
  /** Whether this player is the declarer (always the human, seat 0). */
  isDeclarer: boolean;
}

/** A card played into the current Ulti trick. */
export interface UltiTrickCard {
  playerIdx: number;
  card: Card;
}

/** Ulti game configuration. */
export interface UltiConfig {
  cpuDifficulty: number;
  /** Number of deals that make up the match; the highest coin balance wins. */
  targetRounds: number;
}

/** A suggested hint for Ulti, computed by the backend. */
export interface UltiHint {
  cardIndices: number[];
  /** i18n reason suffix identifier. */
  reason: string;
}

/**
 * Full Ulti (Ultimo) game state returned from the API.
 *
 * Ulti is a 3-player Hungarian contract trick-taker on a 32-card deck
 * (A,10,K,Q,J,9,8,7; trick rank A>10>K>Q>J>9>8>7). The human (seat 0) is
 * always the declarer versus a 2-CPU defending coalition. After the Bid phase
 * (Party / Betli / Durchmarsch, with a trump suit for Party) the declarer takes
 * the 2-card talon and discards 2, then all three play out ten tricks.
 */
export interface UltiResponse extends BaseGameResponse {
  players: UltiPlayer[];
  phase: UltiPhaseValue;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  leadPlayerIdx: number;
  dealerIdx: number;
  /** Seat index of the declarer (always the human, seat 0). */
  declarerIdx: number;
  /** The declared contract (0=None, 1=Party, 2=Betli, 3=Durchmarsch). */
  contract: number;
  /** The trump suit (1=♠ 2=♣ 3=♥ 4=♦), or -1 when none / not a Party contract. */
  trumpSuit: number;
  /** Number of face-down talon cards remaining. */
  talonCount: number;
  /** Whether the declarer has picked up the talon. */
  talonTaken: boolean;
  /** Number of cards discarded so far in the Discard phase. */
  discardCount: number;
  currentTrick: UltiTrickCard[];
  /** Cumulative coin balance per player — [p0, p1, p2]. */
  playerCoins: number[];
  /** 直近ディールの精算による符号付き増減。次のディールが始まると全員 0 に戻る。 */
  lastDealCoins: number[];
  /** Seat index of the last trick winner, or -1. */
  lastTrickWinner: number;
  /** Deal outcome (0=None, 1=Win/contract made, 2=Loss/contract failed). */
  outcome: number;
  /** Match result from the human's perspective (-1 lose, 0 none, 1 win). */
  result: number;
  /** Indices in the human's hand that are legal to play (non-empty on human Play turn). */
  playableIndices: number[];
  gameEndFlag: boolean;
  /** Winning player seat index, or -1 until the game ends. */
  winnerPlayer: number;
  /** Whether it is currently the human's turn to act (play or discard). */
  isHumanTurn: boolean;
  /** Whether it is currently the human's turn to declare a contract. */
  isHumanBidTurn: boolean;
  hint?: UltiHint | null;
  config: UltiConfig;
}

// --- French Tarot ---
