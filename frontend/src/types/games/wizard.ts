// Type declarations for wizard. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Wizard player data with scores. */
export interface WizardPlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  bid: number;
  roundScore: number;
  cumulativeScore: number;
  trickCount: number;
}

/** A card played in a Wizard trick. */
export interface WizardTrickCard {
  playerIdx: number;
  card: Card;
}

/** Wizard game configuration. */
export interface WizardConfig {
  cpuDifficulty: number;
}

/** A suggested hint for Wizard. */
export interface WizardHint {
  cardIndex?: number;
  bid?: number;
  reason: string;
}

/** Full Wizard game state returned from the API. */
export interface WizardResponse extends BaseGameResponse {
  players: WizardPlayerData[];
  phase: number;
  roundNumber: number;
  totalRounds: number;
  handSize: number;
  trickNumber: number;
  currentPlayerIdx: number;
  bidPlayerIdx: number;
  dealerIdx: number;
  currentTrick: WizardTrickCard[];
  trumpCard: Card | null;
  trumpSuit: number;
  restrictedBid: number;
  gameEndFlag: boolean;
  winnerIdx: number;
  leadPlayerIdx: number;
  hint?: WizardHint;
  config: WizardConfig;
}

// --- Ninety-Nine (ナインティナイン) ---
