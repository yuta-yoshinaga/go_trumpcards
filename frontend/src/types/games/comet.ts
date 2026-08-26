// Type declarations for comet. Follows the split-out convention of card.ts
// (issue #4366); card.ts re-exports this file.

import type { BaseGameResponse, Card } from '../common';

/** One seat at the table. */
export interface CometPlayer {
  id: number;
  isHuman: boolean;
  /** Hand cards. Populated only for the human. */
  cards: Card[];
  cardCount: number;
  /** Running match score. */
  score: number;
  isDealer: boolean;
}

/** One round's scoring result. */
export interface CometResult {
  winnerIdx: number;
  cardsLeft: number[];
  gained: number[];
  /** Kings never played this round. Each is worth 2 to the winner. */
  unplayedKings: number;
  /** Seat left holding the Comet, or -1. That seat loses a point. */
  heldWildIdx: number;
}

/** Comet game configuration. */
export interface CometConfig {
  cpuDifficulty: number;
  /** Seats at the table, 2-5. */
  players: number;
  /** Points needed to win, 20-200. */
  targetScore: number;
}

/**
 * Full Comet game state returned from the API.
 *
 * The ancestor of the stops family. **Sequences climb by rank and ignore suit
 * entirely**, the 9 of diamonds is a wild Comet that stands in for any rank,
 * and both it and a king stop the sequence — so the server computes
 * `playableIdxs` rather than leaving the client to re-derive them.
 */
export interface CometResponse extends BaseGameResponse {
  players: CometPlayer[];
  /** "play" | "roundEnd" | "gameEnd". */
  phase: string;
  roundNumber: number;
  dealerIdx: number;
  currentPlayerIdx: number;
  /** Cards played in the current sequence, oldest first. */
  pile: Card[];
  /** The rank needed next, or 0 when this seat leads a fresh sequence. */
  need: number;
  /** Cards buried face down. Sequences stop on ranks that sit here. */
  deadCount: number;
  /** Seat that played last. After a stop, play restarts from here. */
  lastPlayerIdx: number;
  /** Hand indices the human may play. Empty means pass, or not their turn. */
  playableIdxs: number[];
  lastResult: CometResult | null;
  gameEndFlag: boolean;
  winnerIdx: number;
  isHumanTurn: boolean;
  /** Recommended hand card, or -1. */
  hintHandIdx: number;
  hintReason: string;
  config: CometConfig;
}
