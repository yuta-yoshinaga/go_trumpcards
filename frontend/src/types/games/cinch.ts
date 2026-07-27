// Type declarations for cinch. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Cinch game phase (0=Bid 1=NameTrump 2=Play 3=TrickEnd 4=RoundEnd 5=GameEnd). */
export type CinchPhaseValue = 0 | 1 | 2 | 3 | 4 | 5;

/** A Cinch player's public/own state. Cards are non-empty only for the human. */
export interface CinchPlayer {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  trickCount: number;
  /** This player's bid this deal (0=pass, 1-14; -1 if not yet bid). */
  bid: number;
  /** Cumulative match score of this individual player. */
  totalScore: number;
}

/** A card played into the current Cinch trick. */
export interface CinchTrickCard {
  playerIdx: number;
  card: Card;
}

/** Per-deal scoring breakdown for Cinch (surfaced at round/game end). */
export interface CinchDealDetail {
  /** Trump suit for the scored deal (1=♠ 2=♣ 3=♥ 4=♦). */
  trumpSuit: number;
  /** Seat index of the bid winner (bidder) for the deal. */
  bidderIdx: number;
  /** The winning bid amount. */
  bid: number;
  /** Whether the bidding side was set back (failed to make its bid). */
  setBack: boolean;
  /** Points captured per player this deal, keyed by seat index. */
  points: Record<number, number>;
  /** Match points gained (or lost) per player this deal, keyed by seat index. */
  gained: Record<number, number>;
}

/** Cinch game configuration. */
export interface CinchConfig {
  cpuDifficulty: number;
  pointLimit: number;
}

/** A suggested hint for Cinch, computed by the backend. */
export interface CinchHint {
  cardIndices: number[];
  /** Suggested bid amount (present for a bid-phase hint). */
  bid?: number | null;
  /** Suggested trump suit (present for a name-trump-phase hint). */
  trumpSuit?: number | null;
  /** i18n reason suffix identifier. */
  reason: string;
}

/**
 * Full Cinch (Double Pedro / High Five) game state returned from the API.
 *
 * Cinch is a 4-player (1 human + 3 CPU, individual scoring) All-Fours/Pitch-family
 * bidding trick-taker on a standard 52-card deck. Nine cards are dealt to each
 * player; players bid 0 (pass) or 1-14, and the high bidder names trump and leads.
 * There are 14 points per deal — High/King/Ten/Jack of trump = 1 each, the Right
 * Pedro (5 of trump) = 5, and the Left Pedro (5 of the same color as trump, which
 * ranks just below the trump 5) = 5. The bidding side must capture at least its
 * bid or it is set back; the first player to reach the target score (default 21)
 * wins.
 */
export interface CinchResponse extends BaseGameResponse {
  players: CinchPlayer[];
  phase: CinchPhaseValue;
  roundNumber: number;
  trickNumber: number;
  totalTricks: number;
  dealerIdx: number;
  /** Seat index of the player whose turn it is to act. */
  currentTurn: number;
  /** Seat index of the player whose turn it is to bid. */
  bidPlayerIdx: number;
  /** The current highest bid this deal (0 if none yet). */
  currentBid: number;
  /** Seat index of the bid winner, or -1 until decided. */
  bidWinnerIdx: number;
  /** Trump suit (0=unset, 1=♠ 2=♣ 3=♥ 4=♦). */
  trumpSuit: number;
  currentTrick: CinchTrickCard[];
  lastTrick: CinchTrickCard[];
  /** Seat index of the last trick winner, or -1. */
  lastTrickWinner: number;
  /** Indices in the human's hand that are legal to play (non-empty on human Play turn). */
  playableIndices: number[];
  gameEndFlag: boolean;
  /** Winning player seat index, or -1 until the game ends. */
  winnerIdx: number;
  /** Seat indices of players who have reached / won at game end. */
  roundWinners: number[];
  lastDealDetail?: CinchDealDetail | null;
  /** Whether it is currently the human's turn to act. */
  isHumanTurn: boolean;
  hint?: CinchHint | null;
  config: CinchConfig;
}

// --- Loo (Lanterloo) ---
