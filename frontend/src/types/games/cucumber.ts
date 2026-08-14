// Type declarations for cucumber. Split-file layout introduced by issue #4366;
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** A card played into the current Cucumber trick. */
export interface CucumberTrickCard {
  playerIdx: number;
  card: Card;
}

/** One seat at a Cucumber table. */
export interface CucumberPlayer {
  id: number;
  isHuman: boolean;
  cardCount: number;
  /** Populated for the human player only; empty for the CPU. */
  cards: Card[];
  /** Penalty points. **Fewer is better** — this is the only score. */
  penalty: number;
}

/** A suggestion. */
export interface CucumberHint {
  cardIndex?: number;
  /** `cucumberLead`, `cucumberBeat`, or `cucumberForced`. */
  reason: string;
}

/** Table-size and target-score settings. */
export interface CucumberConfig {
  /** Players at the table (3..6, default 4). */
  playerCnt: number;
  /** Penalty total that ends the game (10..100, default 30). */
  targetScore: number;
}

/** Full Cucumber game state returned from the API. */
export interface CucumberResponse extends BaseGameResponse {
  players: CucumberPlayer[];
  /** `0` = Play, `1` = RoundEnd, `2` = GameEnd. */
  phase: number;
  /**
   * Hand indices you may legally play.
   *
   * You must beat `highestInTrick` if you can; if you cannot, this holds
   * exactly your lowest card.
   */
  validPlays: number[];
  /** The rank to beat, or `0` while nobody has led. Suits are irrelevant. */
  highestInTrick: number;
  /**
   * Whether your lowest card is the only legal play.
   *
   * **This is not the same as `validPlays.length === 1`** — a hand can have
   * exactly one card that does beat the trick.
   */
  forced: boolean;
  currentTrick: CucumberTrickCard[];
  currentPlayerIdx: number;
  leadPlayerIdx: number;
  trickNumber: number;
  roundNumber: number;
  /** The seat that took the last trick of the previous round, or `-1`. */
  lastTrickWinnerIdx: number;
  /** Penalty scored for that last trick. */
  lastPenalty: number;
  gameEndFlag: boolean;
  /** `-1` until decided. */
  winnerIdx: number;
  hint?: CucumberHint;
  config: CucumberConfig;
}
