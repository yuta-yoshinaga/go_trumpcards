// Type declarations for russianbank. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** A single Russian Bank (Crapette) player as returned from the API. */
export interface RussianBankPlayer {
  /** Seat index (0 = human). */
  id: number;
  /** True for the human player. */
  isHuman: boolean;
  /** Number of cards left in the reserve (the pile to empty to win). */
  reserveCount: number;
  /** Top reserve card (face up), if any. */
  reserveTop?: Card;
  /** Number of face-down cards left in hand. */
  handCount: number;
  /** Number of cards in the waste pile. */
  wasteCount: number;
  /** Top waste card, if any. */
  wasteTop?: Card;
  /** Number of times this player caught the opponent with "stop". */
  stopPoints: number;
}

/** Russian Bank (Crapette) config echoed back by the server. */
export interface RussianBankConfig {
  cpuDifficulty: number;
}

/** A move hint returned by the Russian Bank /russianbank/exec endpoint. */
export interface RussianBankHint {
  /** Source zone (0=reserve, 1=waste, 2=tableau). */
  zone: number;
  /** True when the source is the opponent's pile. */
  fromOpponent: boolean;
  /** Source tableau column (when zone=tableau). */
  col: number;
  /** True when the destination is a foundation. */
  toFoundation: boolean;
  /** Destination tableau column (when toFoundation is false). */
  toCol: number;
}

/** Server response for the Russian Bank (Crapette) game (POST /russianbank/exec). */
export interface RussianBankResponse extends BaseGameResponse {
  /** Current phase (0=Idle, 1=Playing, 2=GameEnd). */
  phase: number;
  /** Seat index whose turn it currently is. */
  currentPlayerIdx: number;
  /** True when the game has ended. */
  gameEndFlag: boolean;
  /** Winning seat index, or -1 for a draw. */
  winnerIdx: number;
  /** True when it is the human player's turn. */
  isHumanTurn: boolean;
  /** True when the human may call "stop" on the CPU. */
  canCallStop: boolean;
  /** True when the human's last move can be undone. */
  canUndo: boolean;
  /** Total moves played so far. */
  moveCount: number;
  /** The 4 shared tableau columns (top card is last). */
  tableau: Card[][];
  /** The 8 shared foundations (A-up by suit; top card is last). */
  foundations: Card[][];
  /** Optional move hint (present only on a hint request). */
  hint?: RussianBankHint;
  /** One entry per player (index 0 = human). */
  players: RussianBankPlayer[];
  /** Echoed game configuration. */
  config: RussianBankConfig;
}
