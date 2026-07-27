// Type declarations for teenpatti. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Teen Patti game phase (0=Betting 1=SideShow 2=Showdown 3=RoundEnd 4=GameEnd). */
export type TeenPattiPhaseValue = 0 | 1 | 2 | 3 | 4;

/**
 * A Teen Patti player's public/own state. `cards` is populated for the human
 * (once seen) and for everyone at showdown; `handName` is set only when a hand
 * is revealed.
 */
export interface TeenPattiPlayer {
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

/** Teen Patti game configuration. */
export interface TeenPattiConfig {
  cpuDifficulty: number;
  /** Chips put in the pot by each player at the start of a deal. */
  ante: number;
  /** Chips each player begins the match with. */
  startingChips: number;
}

/**
 * A suggested hint for Teen Patti, computed by the backend. `action` is the
 * suggested betting action (e.g. `see`, `bet`, `raise`, `fold`, `show`,
 * `sideshow`).
 */
export interface TeenPattiHint {
  /** Suggested betting action identifier. */
  action: string;
  /** i18n reason suffix identifier. */
  reason: string;
}

/** One participant's revealed hand in a resolved Side Show. */
export interface TeenPattiSideShowHand {
  /** Seat index of this participant. */
  playerIdx: number;
  /** Hand ranking name key (see `hand.*` i18n). */
  handName: string;
  /** The participant's three cards, revealed for the comparison. */
  cards: Card[];
}

/**
 * The comparison result of the most recent accepted Side Show, present only
 * when the human was a participant (CPU-vs-CPU comparisons stay hidden).
 */
export interface TeenPattiSideShowResult {
  /** Seat index that requested the Side Show. */
  requesterIdx: number;
  /** Seat index that accepted the Side Show. */
  targetIdx: number;
  /** Seat index that won the comparison. */
  winnerIdx: number;
  /** Seat index that lost and folded. */
  loserIdx: number;
  /** The requester's revealed hand. */
  requester: TeenPattiSideShowHand;
  /** The target's revealed hand. */
  target: TeenPattiSideShowHand;
}

/**
 * Full Teen Patti game state returned from the API.
 *
 * Teen Patti is the Indian variant of Three Card Brag — a 4-player vying game
 * played with a 52-card deck, 3 cards each, and chips wagered into a pot. Each
 * player is Blind or Seen; on their turn they can See (reveal, Blind→Seen), Bet
 * (call the stake — Blind pays the stake, Seen pays double), Raise, or Fold.
 * When two players remain a Seen player may Show to force a showdown. Teen
 * Patti additionally lets a Seen player request a **Side Show** with the
 * previous Seen player (a private hand comparison; the loser folds), which the
 * target then accepts or declines. The last player standing in a deal wins the
 * pot, chip-busted players are eliminated, and the last player with chips wins
 * the match. Hand ranking is Trail (trio) > Pure Sequence (straight flush) >
 * Sequence (straight) > Color (flush) > Pair > High Card.
 */
export interface TeenPattiResponse extends BaseGameResponse {
  players: TeenPattiPlayer[];
  /** Chips currently in the pot. */
  pot: number;
  /** Current stake a Blind player must match to bet. */
  stake: number;
  phase: TeenPattiPhaseValue;
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
  /** Whether the current player may request a Side Show right now. */
  canRequestSideShow: boolean;
  /** Seat index that requested a Side Show, or -1 when none pending. */
  sideShowRequester: number;
  /** Seat index asked to accept/decline a Side Show, or -1 when none pending. */
  sideShowTarget: number;
  gameEndFlag: boolean;
  /** Whether it is currently the human's turn to act. */
  isHumanTurn: boolean;
  hint?: TeenPattiHint | null;
  /** Result of the last human-involved Side Show comparison, if any. */
  lastSideShow?: TeenPattiSideShowResult | null;
  config: TeenPattiConfig;
}

// --- Préférence ---
