// Type declarations for omi. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Omi player data with team, trick count, and hand. */
export interface OmiPlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  team: number;
  trickCount: number;
}

/** A card played in a Omi trick. */
export interface OmiTrickCard {
  playerIdx: number;
  card: Card;
}

/** Omi game configuration. */
export interface OmiConfig {
  cpuDifficulty: number;
  pointLimit: number;
}

/** A suggested hint for Omi. */
export interface OmiHint {
  cardIndex?: number;
  suit?: number;
  reason: string;
}

/** Full Omi game state returned from the API.
 * Field names match the `json:` tags in OmiWebOutput (OmiWebController.go). */
export interface OmiResponse extends BaseGameResponse {
  players: OmiPlayerData[];
  phase: number;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  /** Index of the player who calls trump (= bidPlayerIdx). */
  trumpCallerIdx: number;
  bidPlayerIdx: number;
  dealerIdx: number;
  trumpSuit: number;
  /** Deal stage: 1 = first 4 cards dealt (CallTrump phase); 2 = all 8 cards dealt (Play onwards). */
  dealStage: number;
  /** Always null in Omi (no face-up card mechanic). */
  faceUpCard: Card | null;
  makerTeam: number;
  /** Always false in Omi (no going-alone mechanic). */
  goingAlone: boolean;
  /** Always -1 in Omi (no going-alone mechanic). */
  goingAlonePlayerIdx: number;
  currentTrick: OmiTrickCard[];
  teamScores: number[];
  /** Tricks won per team this round: [team0Tricks, team1Tricks]. */
  teamTricks: number[];
  gameEndFlag: boolean;
  winnerTeam: number;
  leadPlayerIdx: number;
  config: OmiConfig;
  hint?: OmiHint;
}

// --- Belote (ベロート) ---
