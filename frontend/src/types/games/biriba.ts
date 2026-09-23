// Type declarations for biriba. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Biriba game configuration. */
export interface BiribaConfig {
  cpuDifficulty: number;
  pointLimit: number;
}

/** A single meld on the table in Biriba. */
export interface BiribaMeldData {
  cards: Card[];
  isNatural: boolean;
  isBiriba: boolean;
  rank: number;
}

/** Biriba player data with melds, red 3s, and pozzetto status. */
export interface BiribaPlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  melds: BiribaMeldData[];
  red3Count: number;
  red3s: Card[];
  roundScore: number;
  cumulativeScore: number;
  hasBiriba: boolean;
  hasInitMeld: boolean;
  tookPozzetto: boolean;
}

/** Full Biriba game state returned from the API. */
export interface BiribaResponse extends BaseGameResponse {
  /**
   * The hint the domain computed for the human's turn, when there is one.
   *
   * The CUI has always shown this (which pile to draw from, which cards meld,
   * which discard is safe) with card indices and a reason; the web page used
   * to guess from the phase alone, so the two could disagree (#5628).
   */
  hint?: {
    action: string;
    indices?: number[];
    reason: string;
  };

  players: BiribaPlayerData[];
  phase: number;
  roundNumber: number;
  currentPlayerIdx: number;
  discardTop: Card | null;
  /** The full discard pile, oldest (bottom) first. In Biriba the whole pile is
   * taken at once, so its contents are public information for all players. */
  discardPile: Card[];
  drawPileCount: number;
  discardPileCount: number;
  pozzettoCount: number;
  isFrozen: boolean;
  gameEndFlag: boolean;
  winnerIdx: number;
  config: BiribaConfig;
}
