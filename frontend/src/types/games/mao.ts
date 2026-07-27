// Type declarations for mao. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Mao player data with scores and declaration state. */
export interface MaoPlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  roundScore: number;
  cumulativeScore: number;
  hasDeclared: boolean;
}

/** Mao game configuration. */
export interface MaoConfig {
  cpuDifficulty: number;
  pointLimit: number;
}

/**
 * Full Mao game state returned from the API.
 *
 * The hidden rule is never sent to the client. Only indirect signals are
 * exposed: {@link MaoResponse.awaitingWord} (a word may be required),
 * {@link MaoResponse.rulePenalty} (the last action broke the hidden rule),
 * {@link MaoResponse.correctCount} (successful compliances so far), and
 * {@link MaoResponse.hintUnlocked}/{@link MaoResponse.ruleHint} (a vague hint,
 * populated only after 3 correct compliances).
 */
export interface MaoResponse extends BaseGameResponse {
  players: MaoPlayerData[];
  phase: number;
  roundNumber: number;
  currentPlayerIdx: number;
  discardTop: Card | null;
  drawPileCount: number;
  chosenSuit: number;
  penaltyDrawCount: number;
  direction: number;
  gameEndFlag: boolean;
  winnerIdx: number;
  awaitingWord: boolean;
  correctCount: number;
  hintUnlocked: boolean;
  ruleHint: string;
  rulePenalty: boolean;
  config: MaoConfig;
}

// --- Page One (ページワン) ---
