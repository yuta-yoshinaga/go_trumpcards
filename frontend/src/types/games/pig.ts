// Type declarations for pig. Split-file layout introduced by issue #4366;
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** One seat at a Pig table. */
export interface PigPlayer {
  id: number;
  isHuman: boolean;
  /** Hand size. **Always 4** — everyone passes and receives at the same time. */
  cardCount: number;
  /** Populated for the human player only; empty for the CPU. */
  cards: Card[];
  /** Letters collected. Three puts you out. */
  letters: number;
  /** The letters themselves: `""`, `"P"`, `"PI"`, or `"PIG"`. */
  letterWord: string;
  eliminated: boolean;
  /** Whether this seat has noticed the signal in the current round. */
  hasSignalled: boolean;
  /** The order in which this seat noticed (1-based, `0` = not yet). */
  noticedOrder: number;
  /** Whether this seat has chosen its card to pass. Everyone passes at once. */
  hasChosenPass: boolean;
}

/** A suggestion. Carries no card index while a signal is out. */
export interface PigHint {
  cardIndex?: number;
  /** `pigSignal`, `pigDiscardOdd`, or `pigNoSingleton`. */
  reason: string;
}

/** Table-size and CPU settings. */
export interface PigConfig {
  /** Players at the table (3..6, default 4). The deck is sized to match. */
  playerCnt: number;
  /** `0` = easy, `1` = normal, `2` = hard. How fast the CPU notices a signal. */
  cpuDifficulty: number;
}

/** Full Pig game state returned from the API. */
export interface PigResponse extends BaseGameResponse {
  players: PigPlayer[];
  /** `0` = Pass, `1` = Signal, `2` = RoundEnd, `3` = GameEnd. */
  phase: number;
  /** Hand indices you may pass. Empty once you have chosen this round. */
  validPlays: number[];
  /**
   * The seat that put a hand to its nose, or `-1`.
   *
   * **Nothing on the board shows this** — a signal is silent by definition, so
   * this field is the only way the page can surface it.
   */
  signallerIdx: number;
  /** How many seats have noticed, including the signaller. */
  noticedCnt: number;
  /** The seat that took a letter at the end of the last round, or `-1`. */
  roundLoserIdx: number;
  /**
   * The word letters accumulate towards (`"PIG"`). Three letters and you are out —
   * the page shows it as the denominator of each seat's progress.
   */
  letterTarget: string;
  roundNumber: number;
  /** Passes made this round. */
  passCount: number;
  /** Cards in play: players x 4, so 12 / 16 / 20 / 24. */
  deckSize: number;
  currentPlayerIdx: number;
  gameEndFlag: boolean;
  /** `-1` until decided. */
  winnerIdx: number;
  hint?: PigHint;
  config: PigConfig;
}
