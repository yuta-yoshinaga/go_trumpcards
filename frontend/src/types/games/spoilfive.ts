// Type declarations for spoilfive. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Spoil Five game phase (0=Play 1=TrickEnd 2=RoundEnd 3=GameEnd). */
export type SpoilFivePhaseValue = 0 | 1 | 2 | 3;

/** A Spoil Five player's public/own state. Cards are non-empty only for the human. */
export interface SpoilFivePlayer {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  /** Total tricks taken across the match (cumulative). */
  trickCount: number;
  /** Match points scored so far (first to targetPoints wins). */
  score: number;
  /** Tricks taken in the current round (resets each round; first to 3 takes the pot). */
  roundTricks: number;
}

/** A card played into the current Spoil Five trick. */
export interface SpoilFiveTrickCard {
  playerIdx: number;
  card: Card;
}

/** Spoil Five game configuration. */
export interface SpoilFiveConfig {
  cpuDifficulty: number;
  /** Match points needed to win (default 30). */
  targetPoints: number;
}

/** A suggested hint for Spoil Five, computed by the backend. */
export interface SpoilFiveHint {
  cardIndices: number[];
  /** i18n reason suffix identifier. */
  reason: string;
}

/**
 * Full Spoil Five game state returned from the API.
 *
 * Spoil Five (Maw) is an Irish play-only trick-taker for 5 players on a 52-card
 * deck (5 cards each). Trump is the turned-up card. Fixed top trumps — the trump
 * 5 (highest), trump J, and ♥A (always a trump) — may be held back rather than
 * following suit (Reneging). The first player to win 3 of the 5 tricks takes the
 * pot immediately; if nobody reaches 3 it is a Spoil (流局) and the pot carries
 * to the next round. First player to targetPoints wins the match.
 */
export interface SpoilFiveResponse extends BaseGameResponse {
  players: SpoilFivePlayer[];
  phase: SpoilFivePhaseValue;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  leadPlayerIdx: number;
  dealerIdx: number;
  /** Trump suit (1=♠ 2=♣ 3=♥ 4=♦). */
  trumpSuit: number;
  /** Accumulated pot, awarded to the first player to win 3 tricks. */
  pot: number;
  /** Seat index of the round's winner, or -1 on a Spoil (流局). */
  roundWinnerIdx: number;
  currentTrick: SpoilFiveTrickCard[];
  /** Indices in the human's hand that are legal to play (non-empty on human Play turn). */
  playableIndices: number[];
  gameEndFlag: boolean;
  /** Winning player seat index, or -1 until the game ends. */
  winnerPlayer: number;
  /** Whether it is currently the human's turn to act. */
  isHumanTurn: boolean;
  hint?: SpoilFiveHint | null;
  config: SpoilFiveConfig;
}

// --- Mariáš ---
