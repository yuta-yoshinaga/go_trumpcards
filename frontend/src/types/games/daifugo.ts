// Type declarations for daifugo. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Daifugo player data with rank and card count. */
export interface DaifugoPlayerData {
  id: number;
  isHuman: boolean;
  isFinished: boolean;
  rank: number;
  cardCount: number;
  cards: Card[];
  illegalFinishPenalty?: boolean;
}

/** A play or pass action in Daifugo. */
export interface DaifugoAction {
  playerIdx: number;
  playedCards: Card[] | null; // null = pass
}

/** Daifugo game rule configuration. */
export interface DaifugoConfig {
  jokerCount: number;
  eightCutEnabled: boolean;
  suitLockMode: number;
  elevenBackEnabled: boolean;
  sequenceEnabled: boolean;
  cardExchangeEnabled: boolean;
  blindExchangeEnabled: boolean;
  fiveSkipEnabled: boolean;
  fiveSkipCount: number;
  sevenPassEnabled: boolean;
  tenDiscardEnabled: boolean;
  spadeThreeEnabled: boolean;
  capitalFallEnabled: boolean;
  nineReverseEnabled: boolean;
  coupDetatEnabled: boolean;
  numberLockEnabled: boolean;
  sandstormEnabled: boolean;
  emperorEnabled: boolean;
  sequenceRevolutionEnabled: boolean;
  sequenceLockEnabled: boolean;
  illegalFinishEnabled: boolean;
  queenBomberEnabled: boolean;
  cpuDifficulty: number;
}

/** Input type alias for Daifugo configuration. */
export type DaifugoConfigInput = DaifugoConfig;

/** Card exchange action between ranked players in Daifugo. */
export interface DaifugoExchangeAction {
  fromPlayerIdx: number;
  toPlayerIdx: number;
  cards: Card[];
}

/** Full Daifugo game state returned from the API. */
export interface DaifugoResponse extends BaseGameResponse {
  players: DaifugoPlayerData[];
  currentTurn: number;
  tableCards: Card[];
  lastPlayPlayerIdx: number;
  gameEndFlag: boolean;
  revolutionActive: boolean;
  elevenBackActive: boolean;
  suitLocked: boolean;
  lockedSuit: string;
  tableIsSequence: boolean;
  config: DaifugoConfig;
  exchangeActions: DaifugoExchangeAction[];
  cpuActions: DaifugoAction[];
  humanAction: DaifugoAction | null;
  pendingAction: 'none' | 'sevenPass' | 'tenDiscard' | 'queenBomber';
  pendingActionTarget: number;
  reverseDirection: boolean;
  numberLocked: boolean;
  sequenceLocked: boolean;
  sortMode: number;
  /**
   * Hand indices that appear in at least one legal play right now, mirroring the
   * `*` marks the CUI prints. `null` means the server could not decide (not the
   * human's turn, a pending action, or too many combinations to enumerate) --
   * **not** that nothing is playable, so render no marks rather than marking
   * every card as unplayable.
   */
  playableCardIndices: number[] | null;
}
