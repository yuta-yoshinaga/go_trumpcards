// Type declarations for threecardbrag. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Three Card Brag game phase (0=Betting 1=Showdown 2=RoundEnd 3=GameEnd). */
export type ThreeCardBragPhaseValue = 0 | 1 | 2 | 3;

/**
 * A Three Card Brag player's public/own state. `cards` is populated for the
 * human (once seen) and for everyone at showdown; `handName` is set only when
 * a hand is revealed.
 */
export interface ThreeCardBragPlayer {
  id: number;
  isHuman: boolean;
  /** Remaining chips. */
  chips: number;
  /** Whether the player has looked at their hand (Seen) vs still Blind. */
  seen: boolean;
  /** Whether the player has folded out of the current deal. */
  folded: boolean;
  /** Whether the player has been eliminated (busted) from the match. */
  out: boolean;
  /** Chips this player has wagered into the pot this deal. */
  roundBet: number;
  cardCount: number;
  cards: Card[];
  /** The hand ranking name, set once the hand is revealed. */
  handName?: string;
}

/** Three Card Brag game configuration. */
export interface ThreeCardBragConfig {
  cpuDifficulty: number;
  /** Chips put in the pot by each player at the start of a deal. */
  ante: number;
  /** Chips each player begins the match with. */
  startingChips: number;
}

/**
 * A suggested hint for Three Card Brag, computed by the backend. `action` is
 * the suggested betting action (e.g. `see`, `bet`, `raise`, `fold`, `show`).
 */
export interface ThreeCardBragHint {
  /** Suggested betting action identifier. */
  action: string;
  /** i18n reason suffix identifier. */
  reason: string;
}

/**
 * Full Three Card Brag game state returned from the API.
 *
 * Three Card Brag is a 4-player British vying game (an ancestor of poker)
 * played with a 52-card deck, 3 cards each, and chips wagered into a pot. Each
 * player is Blind or Seen; on their turn they can See (reveal, Blind→Seen), Bet
 * (call the stake — Blind pays the stake, Seen pays double), Raise, or Fold.
 * When two players remain a Seen player may Show to force a showdown. The last
 * player standing in a deal wins the pot, chip-busted players are eliminated,
 * and the last player with chips wins the match. Hand ranking is
 * Prial > Running Flush > Run > Flush > Pair > High Card.
 */
export interface ThreeCardBragResponse extends BaseGameResponse {
  players: ThreeCardBragPlayer[];
  /** Chips currently in the pot. */
  pot: number;
  /** Current stake a Blind player must match to bet. */
  stake: number;
  phase: ThreeCardBragPhaseValue;
  roundNumber: number;
  dealerIdx: number;
  currentPlayerIdx: number;
  /** Winning seat index of the current deal, or -1 until it ends. */
  roundWinnerIdx: number;
  /** Winning seat index of the match, or -1 until the game ends. */
  matchWinnerIdx: number;
  /** Whether the deal has reached a showdown (hands revealed). */
  isShowdown: boolean;
  /** Whether a Seen player may Show (force a showdown) right now. */
  canShow: boolean;
  gameEndFlag: boolean;
  /** Whether it is currently the human's turn to act. */
  isHumanTurn: boolean;
  hint?: ThreeCardBragHint | null;
  config: ThreeCardBragConfig;
}

// --- Teen Patti ---
