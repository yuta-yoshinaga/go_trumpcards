// Type declarations for calabresella. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Calabresella game phase (0=Bid 1=Discard 2=Play 3=TrickEnd 4=RoundEnd 5=GameEnd). */
export type CalabresellaPhaseValue = 0 | 1 | 2 | 3 | 4 | 5;

/** A Calabresella player's public/own state. Cards are non-empty only for the human. */
export interface CalabresellaPlayer {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  trickCount: number;
  /** Cumulative match score of this individual player. */
  score: number;
  /** Whether this player is the round's Soloist (won the bid, plays alone). */
  isSoloist: boolean;
  /** Thirds of a point captured by this player in the current round. */
  roundThirds: number;
}

/** A card played into the current Calabresella trick. */
export interface CalabresellaTrickCard {
  playerIdx: number;
  card: Card;
}

/** Calabresella game configuration. */
export interface CalabresellaConfig {
  cpuDifficulty: number;
  targetPoints: number;
}

/** A suggested hint for Calabresella, computed by the backend. */
export interface CalabresellaHint {
  cardIndices: number[];
  /** i18n reason suffix identifier. */
  reason: string;
}

/**
 * Full Calabresella (Terziglio) game state returned from the API.
 *
 * Calabresella is a Calabrian/Italian 3-player 40-card (Tressette-family)
 * trick-taker with a Bid phase, a monte exchange (discard four) phase, and no
 * trump. One Soloist (bid winner) plays alone against the coalition of the
 * other two and must capture more than half of the 33 thirds to win the round.
 */
export interface CalabresellaResponse extends BaseGameResponse {
  players: CalabresellaPlayer[];
  phase: CalabresellaPhaseValue;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  /** Seat index of the player whose turn it is to bid. */
  currentBidderIdx: number;
  leadPlayerIdx: number;
  dealerIdx: number;
  /** Seat index of the forehand (first to bid / lead). */
  forehandIdx: number;
  /** Seat index of the round's Soloist (bid winner), or -1 until decided. */
  soloistIdx: number;
  /** The winning bid (0=none, 1=chiamo, 2=solo). */
  winningBid: number;
  currentTrick: CalabresellaTrickCard[];
  /**
   * The four monte (widow) cards, revealed to every player once the Soloist has
   * taken them (present from the Discard phase onward; omitted during Bidding).
   */
  monte?: Card[];
  /** Cumulative match scores per player — [p0, p1, p2]. */
  playerScores: number[];
  /** Thirds of a point captured per player this round — [p0, p1, p2]. */
  roundThirds: number[];
  /** Seat index of the last trick winner, or -1. */
  lastTrickWinner: number;
  /** Indices in the human's hand that are legal to play (non-empty on human Play turn). */
  playableIndices: number[];
  gameEndFlag: boolean;
  /** Winning player seat index, or -1 until the game ends. */
  winnerPlayer: number;
  /** Whether it is currently the human's turn to act. */
  isHumanTurn: boolean;
  hint?: CalabresellaHint | null;
  config: CalabresellaConfig;
}

// --- Ombre (Hombre) ---
