// Type declarations for hokm. Split-file layout introduced by issue
// #4366; card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** A card played into the current Hokm trick. */
export interface HokmTrickCard {
  playerIdx: number;
  card: Card;
}

/** One seat at a Hokm table. */
export interface HokmPlayer {
  id: number;
  isHuman: boolean;
  /** `0` or `1`. Seats 0+2 are one partnership, 1+3 the other. */
  team: number;
  /** Hand size. The only hand information exposed for the CPU. */
  cardCount: number;
  /** Populated for the human player only; empty for the CPU. */
  cards: Card[];
  /** Declares trump from their first five cards; keeps the role while winning. */
  isHakem: boolean;
  trickCount: number;
}

/**
 * A suggestion. While trump is undeclared it carries no `cardIndex` and puts
 * the recommended suit in `suit`; during play it names a card.
 */
export interface HokmHint {
  cardIndex?: number;
  /** `hokmDeclareTrump` before play; `hokmWinTrick` or `hokmSaveCards` during. */
  reason: string;
  /** Suit to make trump; `0` during play. */
  suit: number;
}

/** Target-hands setting. */
export interface HokmConfig {
  /** Hand points needed to win (1..13, default 7). */
  target: number;
}

/** Full Hokm game state returned from the API. */
export interface HokmResponse extends BaseGameResponse {
  players: HokmPlayer[];
  /** `0` = Trump, `1` = Play, `2` = HandEnd, `3` = GameEnd. */
  phase: number;
  handNumber: number;
  trickNumber: number;
  /** `0` until the hakem has declared. */
  trumpSuit: number;
  hakemIdx: number;
  /** Hand points per team, index 0 and 1. */
  scores: number[];
  /**
   * Tricks taken this hand per team. **The hand ends the moment one side
   * reaches `tricksToWin`**, so this — not `trickNumber` — is the progress bar.
   */
  teamTricks: number[];
  /** Tricks that take the hand (7). */
  tricksToWin: number;
  /** The previous hand was a Kot (the losers took no trick), worth 2. */
  lastHandKot: boolean;
  /**
   * Whether the hakem passed to the next seat after the hand just settled.
   *
   * The hakem only moves when its team loses, and which seat calls the trump
   * next is worth knowing before the next hand starts (#5753).
   */
  lastHandHakemChanged: boolean;
  /** Team that took the previous hand, or `-1` before there is one. */
  lastHandWinner: number;
  currentPlayerIdx: number;
  leadPlayerIdx: number;
  currentTrick: HokmTrickCard[];
  /** Hand indices you may legally play. Following suit is compulsory. */
  validPlays: number[];
  gameEndFlag: boolean;
  /** `-1` until decided, and `-1` on a tie. */
  winnerTeam: number;
  hint?: HokmHint;
  config: HokmConfig;
}
