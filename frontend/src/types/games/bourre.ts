// Type declarations for bourre. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Bourré player data. */
export interface BourrePlayerData {
  id: number;
  isHuman: boolean;
  isFinished: boolean;
  folded: boolean;
  decided: boolean;
  drawn: boolean;
  bourreed: boolean;
  chips: number;
  tricks: number;
  cardCount: number;
  cards: Card[];
}

/** A single card played into a Bourré trick. */
export interface BourreTrickCardData {
  playerIdx: number;
  card: Card | null;
}

/** Bourré hand result for one player. */
export interface BourreResultData {
  playerIdx: number;
  tricks: number;
  wonAmount: number;
  bourreed: boolean;
  folded: boolean;
}

/** Bourré config. */
export interface BourreConfig {
  cpuDifficulty: number;
}

/** Bourré API response. */
export interface BourreResponse extends BaseGameResponse {
  players: BourrePlayerData[];
  phase: string;
  currentPlayerIdx: number;
  dealerIdx: number;
  pot: number;
  carryPot: number;
  trumpSuit: string;
  trumpCard: Card | null;
  trickNumber: number;
  currentTrick: BourreTrickCardData[];
  lastTrick: BourreTrickCardData[];
  lastTrickWinner: number;
  leadPlayerIdx: number;
  handNumber: number;
  gameEndFlag: boolean;
  winnerIdx: number;
  validPlays: number[];
  results: BourreResultData[];
  config: BourreConfig;
}

// --- Spoons ---
