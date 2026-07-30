// Type declarations for toepen. Split-file layout introduced by issue
// #4366; card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** One card in the current trick. */
export interface ToepenTrickCard {
  playerIdx: number;
  card: Card;
}

/** One Toepen seat. */
export interface ToepenPlayer {
  id: number;
  isHuman: boolean;
  /** Hand size. Always sent, including while {@link ToepenPlayer.hidden} is true. */
  cardCount: number;
  /** Empty while {@link ToepenPlayer.hidden} is true. */
  cards: Card[];
  /** Lives lost so far; `maxLives` ends this player's game. */
  lives: number;
  /** Out of THIS hand, having declined a toep. */
  folded: boolean;
  /** Out of the game. */
  eliminated: boolean;
  hidden: boolean;
}

/** Suggested move for the human seat. */
export interface ToepenHintPayload {
  cardIndex?: number;
  fold?: boolean;
  /** Reason identifier, e.g. `toepen.hint.play`. */
  reason: string;
}

/** Full Toepen game state returned from the API. */
export interface ToepenResponse extends BaseGameResponse {
  players: ToepenPlayer[];
  /** 0 = Play, 1 = Respond, 2 = HandEnd, 3 = GameEnd. */
  phase: number;
  currentPlayerIdx: number;
  leadPlayerIdx: number;
  dealerIdx: number;
  currentTrick: ToepenTrickCard[];
  /** -1 before anything is led. */
  leadSuit: number;
  trickNumber: number;
  handNumber: number;
  /**
   * Lives this hand now costs. Starts at 1 and rises by one per toep. Folding
   * costs `stake - 1` — the value BEFORE the raise being declined.
   */
  stake: number;
  /** Who toeped; -1 outside a response phase. */
  knockerIdx: number;
  /** Whose answer is awaited; -1 when none. */
  pendingRespondent: number;
  /** The only seat that escapes the penalty when the hand settles. */
  lastTrickWinner: number;
  maxLives: number;
  /**
   * Hand indices the human may legally play. Carries the follow-suit
   * obligation, so the page never re-derives it.
   */
  validPlayIndices: number[];
  gameEndFlag: boolean;
  winnerIdx: number;
  hint?: ToepenHintPayload;
}
