// Type declarations for piquet. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Piquet game configuration. */
export interface PiquetConfig {
  cpuDifficulty: number;
  dealsPerPartie: number;
}

/** Piquet player data. */
export interface PiquetPlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  trickCount: number;
  declScore: number;
  trickScore: number;
  bonusScore: number;
  roundScore: number;
  matchScore: number;
}

/** Piquet trick card data. */
export interface PiquetTrickCard {
  playerIdx: number;
  card: Card;
}

/** Piquet claim (Point/Sequence/Set declaration evidence). */
export interface PiquetClaim {
  length: number;
  topRank: number;
  pipTotal: number;
  suit: number;
  cards: Card[];
}

/** Piquet declaration result. */
export interface PiquetDeclaration {
  kind: number;
  elderClaim?: PiquetClaim;
  youngerClaim?: PiquetClaim;
  winner: number;
  score: number;
  scoredBy: number;
  sets?: PiquetClaim[];
}

/** Full Piquet game state returned from the API. */
export interface PiquetResponse extends BaseGameResponse {
  players: PiquetPlayerData[];
  phase: number;
  dealNumber: number;
  dealsPerPartie: number;
  elderIdx: number;
  youngerIdx: number;
  currentPlayerIdx: number;
  leadPlayerIdx: number;
  trickNumber: number;
  tricksWon: [number, number];
  exchangeTurn: number;
  elderExchangedCnt: number;
  youngerExchangedCnt: number;
  elderTalon: Card[];
  youngerTalon: Card[];
  elderRevealedTalon: Card[];
  youngerRevealedTalon: Card[];
  carteBlanche: [boolean, boolean];
  declStage: number;
  declResults: PiquetDeclaration[];
  currentTrick: PiquetTrickCard[];
  legalPlayIndices?: number[];
  gameEndFlag: boolean;
  winnerIdx: number;
  hint?: {
    cardIndex?: number;
    discardIndices?: number[];
    reason: string;
  };
  config: PiquetConfig;
}

// --- Golf Solitaire (ゴルフ) ---
