// Type declarations for watten. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** A Watten player's public/own state. Cards are non-empty only for the human during play. */
export interface WattenPlayer {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  team: number;
  trickCount: number;
}

/** A card played into the current Watten trick. */
export interface WattenTrickCard {
  playerIdx: number;
  card: Card;
}

/** Watten game configuration. */
export interface WattenConfig {
  cpuDifficulty: number;
  targetScore: number;
  maxRaises: number;
}

/** A suggested hint for Watten, computed by the backend. */
export interface WattenHint {
  /** The suggested action: declare, raise, play, hold, or fold. */
  action: string;
  cardIndex?: number;
  rank?: number;
  suit?: number;
  reason: string;
}

/**
 * Full Watten (ヴァッテン) game state returned from the API.
 *
 * Watten is a Bavarian/Austrian 4-player, 2-team trick-taker with a raise/bluff
 * stake mechanic. Seats 0 & 2 form team 0, seats 1 & 3 form team 1; the human is
 * seat 0. The dealer declares a Schlag rank and a critical (trump) suit, teams
 * play five tricks, and either team may raise the stake for the other to hold or
 * fold. Winning at least three tricks scores the stake; first team to the target
 * score wins.
 */
export interface WattenResponse extends BaseGameResponse {
  players: WattenPlayer[];
  phase: number;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  dealerIdx: number;
  leadPlayerIdx: number;
  /** The declared Schlag rank (1=A, 7..13), or 0 when unset. */
  schlagRank: number;
  /** The declared critical (trump) suit (1=♠ 2=♣ 3=♥ 4=♦), or 0 when unset. */
  criticalSuit: number;
  /** The current accepted stake (starts at 2). */
  stake: number;
  /** The proposed stake after a raise, awaiting a hold/fold response (0 when none). */
  pendingStake: number;
  raiseCount: number;
  /** The team that proposed the pending raise, or -1 when none. */
  raiserTeam: number;
  /** Seat index of the player who must hold/fold a pending raise, or -1 when none. */
  responderIdx: number;
  /** Whether the human (as lead) may raise the stake right now. */
  canRaise: boolean;
  currentTrick: WattenTrickCard[];
  teamScores: number[];
  teamTricks: number[];
  /** The team that won the most recent completed deal, or -1 until decided. */
  dealWinnerTeam: number;
  gameEndFlag: boolean;
  winnerTeam: number;
  /** Match result from the human's (team 0) perspective: -1 lose, 0 none, 1 win. */
  result: number;
  hint?: WattenHint | null;
  config: WattenConfig;
}

// --- Gaigel (ガイゲル) ---
