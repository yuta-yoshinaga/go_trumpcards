// Type declarations for literature. Split-file layout introduced by issue #4366;
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** One Literature seat. */
export interface LiteraturePlayer {
  id: number;
  isHuman: boolean;
  /** 0 for seats 0/2/4, 1 for seats 1/3/5 — **seating alternates**. */
  team: number;
  cardCount: number;
  /**
   * Your own hand only, until the game ends. **Even a teammate's hand stays
   * hidden** — deducing where the cards are is the whole game.
   */
  cards: Card[];
  isCurrentTurn: boolean;
}

/** One entry in the ask history. **Everyone sees these.** */
export interface LiteratureAsk {
  from: number;
  to: number;
  card: Card | null;
  success: boolean;
}

/** One entry in the claim history. */
export interface LiteratureClaim {
  player: number;
  halfSuit: number;
  /**
   * 0 = won, 1 = **cancelled**, 2 = lost to the opponents. Cancelled is *not*
   * the same as losing it — naming the wrong teammate voids the half-suit
   * rather than handing it over.
   */
  outcome: number;
  /** The team that took it; -1 when cancelled. */
  awardedTeam: number;
}

/** Full Literature game state returned from the API. */
export interface LiteratureResponse extends BaseGameResponse {
  players: LiteraturePlayer[];
  /** 0 = Play, 1 = GameEnd. */
  phase: number;
  currentPlayerIdx: number;
  /** Per half-suit: 0 = open, 1 = team 0, 2 = team 1, 3 = **cancelled**. */
  halfSuits: number[];
  /** The six cards of each half-suit, so the UI can offer them. */
  halfSuitCards: Card[][];
  asks: LiteratureAsk[];
  claims: LiteratureClaim[];
  lastAsk: LiteratureAsk | null;
  lastClaim: LiteratureClaim | null;
  teamHalfSuits: [number, number];
  /** **Cancelled half-suits belong to nobody**, so the totals need not reach 8. */
  cancelledCount: number;
  openCount: number;
  /** Half-suits needed to win — **five**, a majority of eight, not four. */
  winThreshold: number;
  halfSuitCnt: number;
  gameEndFlag: boolean;
  /** Winning team; -1 while live, and also -1 if it ends level. */
  winnerTeam: number;
  config: LiteratureConfigOutput;
}

/** Settings echoed back with the game state. */
export interface LiteratureConfigOutput {
  cpuDifficulty: number;
}
