// Type declarations for euchre. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Euchre player data with team, trick count, and hand. */
export interface EuchrePlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  team: number;
  trickCount: number;
}

/** A card played in a Euchre trick. */
export interface EuchreTrickCard {
  playerIdx: number;
  card: Card;
}

/** Euchre game configuration. */
export interface EuchreConfig {
  cpuDifficulty: number;
  pointLimit: number;
}

/** A suggested hint for Euchre. */
export interface EuchreHint {
  cardIndex?: number;
  orderUp?: boolean;
  suit?: number;
  goAlone?: boolean;
  reason: string;
  /** Hand strength the bid decision was made on. Bidding phase only. */
  score?: number;
  /** Thresholds `score` is read against. Sent by the server rather than
   * duplicated here, so the displayed basis cannot drift from the rule. */
  orderUpScore?: number;
  goAloneScore?: number;
}

/** Full Euchre game state returned from the API. */
export interface EuchreResponse extends BaseGameResponse {
  players: EuchrePlayerData[];
  phase: number;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  bidPlayerIdx: number;
  dealerIdx: number;
  trumpSuit: number;
  faceUpCard: Card | null;
  makerTeam: number;
  goingAlone: boolean;
  goingAlonePlayerIdx: number;
  currentTrick: EuchreTrickCard[];
  teamScores: number[];
  gameEndFlag: boolean;
  winnerTeam: number;
  leadPlayerIdx: number;
  config: EuchreConfig;
  hint?: EuchreHint;
}

// --- Belote (ベロート) ---
