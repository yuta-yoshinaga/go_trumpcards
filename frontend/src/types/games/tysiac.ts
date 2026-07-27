// Type declarations for tysiac. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Tysiąc game phase (0=Bid 1=Talon 2=Play 3=TrickEnd 4=RoundEnd 5=GameEnd). */
export type TysiacPhaseValue = 0 | 1 | 2 | 3 | 4 | 5;

/** A Tysiąc player's public/own state. Cards are non-empty only for the human. */
export interface TysiacPlayer {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  trickCount: number;
  /** Cumulative match score of this individual player. */
  score: number;
  /** Whether this player is the round's Declarer (won the bid, plays the contract). */
  isDeclarer: boolean;
}

/** A card played into the current Tysiąc trick. */
export interface TysiacTrickCard {
  playerIdx: number;
  card: Card;
}

/** Tysiąc game configuration. */
export interface TysiacConfig {
  cpuDifficulty: number;
  targetPoints: number;
}

/** A suggested hint for Tysiąc, computed by the backend. */
export interface TysiacHint {
  cardIndices: number[];
  /** i18n reason suffix identifier. */
  reason: string;
}

/**
 * Full Tysiąc (Thousand) game state returned from the API.
 *
 * Tysiąc is a Polish 3-player 24-card trick-taker with a Bid phase, a Talon
 * exchange phase, and marriage (K+Q) declarations during play. The Declarer
 * wins the bid and tries to meet the contract; trump is set dynamically by
 * declaring a marriage (so `trumpSuit` starts at 0 = unset).
 */
export interface TysiacResponse extends BaseGameResponse {
  players: TysiacPlayer[];
  phase: TysiacPhaseValue;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  leadPlayerIdx: number;
  dealerIdx: number;
  /** Seat index of the forehand (first to bid / lead). */
  forehandIdx: number;
  /** Seat index of the round's Declarer (bid winner). */
  declarerIdx: number;
  /** The Declarer's contract (target card points for the round). */
  contract: number;
  /** The current highest bid in the Bid phase. */
  currentBid: number;
  /** Trump suit (0=unset, 1=♠ 2=♣ 3=♥ 4=♦). 0 until a marriage is declared. */
  trumpSuit: number;
  currentTrick: TysiacTrickCard[];
  /** Cumulative match scores per player — [p0, p1, p2]. */
  playerScores: number[];
  /** Card points captured per player this round — [p0, p1, p2]. */
  roundCardPoints: number[];
  /** Marriage (K+Q same suit) points scored per player this round — [p0, p1, p2]. */
  roundMarriage: number[];
  /** Seat index of the last trick winner, or -1. */
  lastTrickWinner: number;
  /** Indices in the human's hand that are legal to play (non-empty on human Play turn). */
  playableIndices: number[];
  gameEndFlag: boolean;
  /** Winning player seat index, or -1 until the game ends. */
  winnerPlayer: number;
  /** Whether it is currently the human's turn to act. */
  isHumanTurn: boolean;
  hint?: TysiacHint | null;
  config: TysiacConfig;
}

// --- Calabresella (Terziglio) ---
