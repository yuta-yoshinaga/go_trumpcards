// Type declarations for michigan. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Michigan game phase (0=Bet, 1=Play, 2=Result). */
export type MichiganPhaseValue = 0 | 1 | 2;

/**
 * A Michigan player's public/own state. `cards` is populated for the human
 * during the play phase and revealed for every player at the result phase; CPU
 * hands are empty (`cards: []`, use `cardCount`) while the round is in progress.
 */
export interface MichiganPlayer {
  id: number;
  isHuman: boolean;
  /** Remaining chips. */
  chips: number;
  /** Chips this player has wagered across the boodles this round. */
  roundBet: number;
  cardCount: number;
  cards: Card[];
  /** Whether it is this player's turn to act. */
  isCurrent: boolean;
  /** Whether this player emptied their hand to end the round. */
  isWinner: boolean;
}

/**
 * A Michigan boodle — one of the four center "betting" cards (A♥, K♣, Q♦, J♠)
 * onto which players stake chips. When a player plays a card matching the
 * boodle they collect its chips.
 */
export interface MichiganBoodle {
  /** The fixed boodle card (A♥, K♣, Q♦, or J♠). */
  card: Card;
  /** Chips currently staked on this boodle. */
  chips: number;
  /** Seat index of the player who claimed the boodle's chips, or -1 if unclaimed. */
  claimedBy: number;
}

/** Michigan local-rule configuration. */
export interface MichiganConfig {
  /** Number of players at the table (3–8). */
  playerCount: number;
  /** Total chips each player distributes across the four boodles per round. */
  ante: number;
  /** Chips each player begins the match with. */
  startingChips: number;
  /** Number of rounds after which the richest player wins the match. */
  targetRounds: number;
}

/**
 * A suggested hint for Michigan, computed by the backend. `cardIndex` is the
 * hand index to play and `reason` is an i18n reason suffix (`forced` /
 * `claim_boodle` / `lead_low`).
 */
export interface MichiganHint {
  /** Hand index of the suggested card to play. */
  cardIndex: number;
  /** i18n reason suffix identifier. */
  reason: string;
}

/**
 * Full Michigan game state returned from the API.
 *
 * Michigan (Newmarket) is a "stops" chip-betting game. Players first stake chips
 * across four center "boodle" cards (A♥, K♣, Q♦, J♠), then play cards in
 * ascending same-suit sequences. Playing a card that matches a boodle wins its
 * chips; emptying your hand ends the round. After `targetRounds` rounds the
 * richest player wins the match.
 */
export interface MichiganResponse extends BaseGameResponse {
  players: MichiganPlayer[];
  boodles: MichiganBoodle[];
  /** Game phase: 0=Bet, 1=Play, 2=Result. */
  phase: MichiganPhaseValue;
  roundNumber: number;
  /** Total chips each player stakes across the boodles per round. */
  ante: number;
  /** The human's remaining chip stack. */
  chips: number;
  /** Chips the human must distribute across the four boodles this round. */
  betBudget: number;
  /** Whether the human has already placed their boodle bets this round. */
  humanBetPlaced: boolean;
  /** Seat index of the player to act. */
  currentPlayerIdx: number;
  /** Seat index of the dealer. */
  dealerIdx: number;
  /** Seat index of the player who leads the current sequence. */
  leadPlayerIdx: number;
  /** Suit of the active sequence (0=none/new sequence needed, 1–4=suit). */
  seqSuit: number;
  /** Display name of the active sequence's suit, or empty when a new one is needed. */
  seqSuitName: string;
  /** Highest card value played so far in the current run. */
  seqHighValue: number;
  /** Whether the current player must start a fresh sequence. */
  needNewSequence: boolean;
  /** Number of cards in the face-down dead hand. */
  deadHandCount: number;
  /** Whether it is the human's turn to act. */
  isHumanTurn: boolean;
  /** Legal hand indices the human may play this turn. */
  playableIndices: number[];
  /** Seat index of the player who emptied their hand, or -1 for none. */
  winnerIdx: number;
  /** Winning seat index of the match, or -1 until it is decided. */
  matchWinnerIdx: number;
  /** The human's round result: 1=win, 0=none, -1=lose. */
  result: number;
  gameEndFlag: boolean;
  hint?: MichiganHint | null;
  config: MichiganConfig;
}
