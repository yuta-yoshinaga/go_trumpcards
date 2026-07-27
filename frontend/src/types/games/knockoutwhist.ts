// Type declarations for knockoutwhist. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Knockout Whist game phase (0=Play 1=TrickEnd 2=RoundEnd 3=GameEnd 4=TrumpSelect). */
export type KnockoutWhistPhaseValue = 0 | 1 | 2 | 3 | 4;

/** A Knockout Whist player's public/own state. Cards are non-empty only for the human. */
export interface KnockoutWhistPlayer {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  /** Total tricks taken across the match (cumulative). */
  trickCount: number;
  /** Whether this player has been knocked out of the match. */
  eliminated: boolean;
  /** Remaining Dogbone survival tokens (each player starts with 1). */
  dogbones: number;
  /** Tricks taken in the current round (resets each round). */
  roundTricks: number;
}

/** A card played into the current Knockout Whist trick. */
export interface KnockoutWhistTrickCard {
  playerIdx: number;
  card: Card;
}

/** Knockout Whist game configuration (CPU difficulty only — no target points). */
export interface KnockoutWhistConfig {
  cpuDifficulty: number;
}

/** A suggested hint for Knockout Whist, computed by the backend. */
export interface KnockoutWhistHint {
  cardIndices: number[];
  /** i18n reason suffix identifier. */
  reason: string;
}

/**
 * Full Knockout Whist game state returned from the API.
 *
 * Knockout Whist is a British play-only survival trick-taker: each round the
 * hand shrinks by one card, the previous round's winner's longest suit becomes
 * trump (auto), and a player who wins zero tricks must spend a Dogbone to
 * survive — or is eliminated. Last player standing wins.
 */
export interface KnockoutWhistResponse extends BaseGameResponse {
  players: KnockoutWhistPlayer[];
  phase: KnockoutWhistPhaseValue;
  roundNumber: number;
  /** Number of cards dealt this round (8 - roundNumber, down to 1). */
  handSize: number;
  trickNumber: number;
  currentPlayerIdx: number;
  leadPlayerIdx: number;
  dealerIdx: number;
  /** Trump suit (1=♠ 2=♣ 3=♥ 4=♦). */
  trumpSuit: number;
  /** Seat index of the round's winner, or -1. */
  roundWinnerIdx: number;
  currentTrick: KnockoutWhistTrickCard[];
  /** Number of players still in the match (not eliminated). */
  activeCount: number;
  /** Indices in the human's hand that are legal to play (non-empty on human Play turn). */
  playableIndices: number[];
  gameEndFlag: boolean;
  /** Winning player seat index, or -1 until the game ends. */
  winnerPlayer: number;
  /** Whether it is currently the human's turn to act. */
  isHumanTurn: boolean;
  hint?: KnockoutWhistHint | null;
  config: KnockoutWhistConfig;
}

// --- Spoil Five ---
