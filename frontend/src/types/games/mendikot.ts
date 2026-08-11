// Type declarations for mendikot. Split-file layout introduced by issue
// #4366; card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** A card played into the current Mendikot trick. */
export interface MendikotTrickCard {
  playerIdx: number;
  card: Card;
}

/** One seat at a Mendikot table. */
export interface MendikotPlayer {
  id: number;
  isHuman: boolean;
  /** `0` or `1`. Seats 0+2 are one partnership, 1+3 the other. */
  team: number;
  /** Hand size. The only hand information exposed for the CPU. */
  cardCount: number;
  /** Populated for the human player only; empty for the CPU. */
  cards: Card[];
  /** Tens captured by this seat. The four tens decide the hand. */
  tens: number;
  trickCount: number;
}

/** A suggestion naming a card to play. */
export interface MendikotHint {
  cardIndex?: number;
  /**
   * `mendikotChaseTen` when a ten is on the table, `mendikotFeedPartner` when
   * your partner already has the trick, `mendikotDuck` when you cannot win it.
   */
  reason: string;
}

/** Target-score setting. */
export interface MendikotConfig {
  /** Hand points needed to win (1..20, default 3). */
  target: number;
}

/** Full Mendikot game state returned from the API. */
export interface MendikotResponse extends BaseGameResponse {
  players: MendikotPlayer[];
  /** `0` = Play, `1` = HandEnd, `2` = GameEnd. There is no trump phase. */
  phase: number;
  handNumber: number;
  trickNumber: number;
  /**
   * `0` until somebody fails to follow suit — the card they play sets it.
   * There is no separate trump-choosing step.
   */
  trumpSuit: number;
  /** Seat whose unfollowable card set trump, or `-1` while it is undecided. */
  trumpChooserIdx: number;
  /** How many tens exist in the deck (4). The hand is a race for these. */
  tensInDeck: number;
  /** Tens captured this hand per team, index 0 and 1. */
  teamTens: number[];
  /** Tricks taken this hand per team, index 0 and 1. Only decides a 2-2 split. */
  teamTricks: number[];
  /** Hand points per team, index 0 and 1. */
  scores: number[];
  /** Team that took the previous hand, or `-1` before the first one ends. */
  lastHandWinner: number;
  /**
   * How the previous hand was decided: `tens`, `tricks`, `mendikot` (all four
   * tens, 2 points) or `whitewash` (every trick, 3 points). Empty before then.
   */
  lastHandKind: string;
  currentPlayerIdx: number;
  leadPlayerIdx: number;
  dealerIdx: number;
  currentTrick: MendikotTrickCard[];
  /** Hand indices you may legally play. Following suit is compulsory. */
  validPlays: number[];
  gameEndFlag: boolean;
  /** `-1` until decided, and `-1` on a tie. */
  winnerTeam: number;
  hint?: MendikotHint;
  config: MendikotConfig;
}
