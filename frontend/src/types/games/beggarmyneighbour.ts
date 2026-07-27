// Type declarations for beggarmyneighbour. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Player data for a Beggar-My-Neighbour participant. */
export interface BeggarMyNeighbourPlayerData {
  /** Player index (0 = human, 1 = CPU). */
  id: number;
  /** Whether this player is the human. */
  isHuman: boolean;
  /** Number of cards in the face-down draw pile. */
  drawPileSize: number;
  /** Number of cards in the discard pile (refills draw pile when empty). */
  discardPileSize: number;
  /** Total cards held (drawPileSize + discardPileSize). */
  totalCards: number;
}

/** Beggar-My-Neighbour game configuration. */
export interface BeggarMyNeighbourConfig {
  /** Maximum rounds before the game is decided by card count. */
  maxRounds: number;
}

/** Full Beggar-My-Neighbour game state returned from the API. */
export interface BeggarMyNeighbourResponse extends BaseGameResponse {
  players: BeggarMyNeighbourPlayerData[];
  /** Current game phase (0=Play, 1=PayPenalty, 2=Collect, 3=GameEnd). */
  phase: number;
  gameEndFlag: boolean;
  winnerIdx: number;
  /** Index of the player whose turn it is (0=human, 1=CPU). */
  currentPlayerIdx: number;
  /** Index of the player who played the last penalty card (-1 if none). */
  penaltyOwnerIdx: number;
  /** Number of penalty cards still to be paid. */
  penaltyRemaining: number;
  /** Number of cards in the central pile. */
  centralPileSize: number;
  /** The last card played onto the central pile, or null. */
  lastCardPlayed: Card | null;
  /** Number of collection rounds completed. */
  roundsPlayed: number;
  config: BeggarMyNeighbourConfig;
}

// --- All Fours (Seven Up / Old Sledge) ---
