// Type declarations for tressette. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** A Tressette player's public/own state. */
export interface TressettePlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  trickCount: number;
  teamId: number;
}

/** A card played in a Tressette trick. */
export interface TressetteTrickCard {
  playerIdx: number;
  card: Card;
}

/** Tressette game configuration. */
export interface TressetteConfig {
  cpuDifficulty: number;
  targetPoints: number;
}

/** A suggested hint for Tressette. */
export interface TressetteHint {
  cardIndices: number[];
  reason: string;
}

/** Full Tressette game state returned from the API. */
export interface TressetteResponse extends BaseGameResponse {
  players: TressettePlayerData[];
  phase: number;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  currentTrick: TressetteTrickCard[];
  lastTrick: TressetteTrickCard[];
  lastTrickWinner: number;
  leadPlayerIdx: number;
  teamScores: number[];
  teamRoundThirds: number[];
  playableIndices: number[];
  gameEndFlag: boolean;
  winnerTeam: number;
  config: TressetteConfig;
  hint?: TressetteHint;
}

// --- Sheepshead ---
