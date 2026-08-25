// Type declarations for diloti. Follows the split-out convention of card.ts
// (issue #4366); card.ts re-exports this file.

import type { BaseGameResponse, Card } from '../common';

/** One seat at the two-handed table. */
export interface DilotiPlayer {
  id: number;
  isHuman: boolean;
  /** Hand cards. Populated only for the human. */
  cards: Card[];
  cardCount: number;
  /** Cards taken this round. */
  capturedCount: number;
  /** Xeri (one-card table sweeps) this round. Each is worth 10. */
  xeri: number;
  /** Running match score. */
  score: number;
  isDealer: boolean;
}

/** A pile declared on the table. */
export interface DilotiDeclaration {
  /** Seat that owes the capture. */
  ownerIdx: number;
  /** The value that takes it. */
  value: number;
  /** Each group totals `value`. A group declaration has two or more. */
  groups: Card[][];
  /** A group declaration cannot be raised and is taken whole or not at all. */
  isGroup: boolean;
}

/** One legal capture: loose table cards and/or declarations. */
export interface DilotiTake {
  tableIdxs: number[];
  declIdxs: number[];
}

/** One declaration a card could make. */
export interface DilotiDeclCandidate {
  value: number;
  tableIdxs: number[];
}

/** One scoring category of a round. */
export interface DilotiScoreLine {
  /** "cards" | "aces" | "tenOfDiamonds" | "twoOfClubs" | "xeri". */
  key: string;
  /** Points per seat. */
  points: number[];
}

/** One round's scoring result. */
export interface DilotiResult {
  lines: DilotiScoreLine[];
  totals: number[];
  cardCounts: number[];
  xeris: number[];
}

/** Diloti game configuration. */
export interface DilotiConfig {
  cpuDifficulty: number;
  /** Points needed to win, 21-101. */
  targetScore: number;
}

/**
 * Full Diloti game state returned from the API.
 *
 * A Greek fishing game. **Captures sum to the played card's own rank, not to
 * ten**, face cards take exactly one match and never join a sum, and declaring
 * builds piles that only their own value can take — so the server enumerates
 * the legal moves rather than leaving the client to re-derive rules that
 * interact.
 */
export interface DilotiResponse extends BaseGameResponse {
  players: DilotiPlayer[];
  /** "play" | "roundEnd" | "gameEnd". */
  phase: string;
  roundNumber: number;
  dealerIdx: number;
  currentPlayerIdx: number;
  /** Loose cards on the table, in the order the indices refer to. */
  table: Card[];
  declarations: DilotiDeclaration[];
  deckRemaining: number;
  /** Seat that captured last. The leftover table goes here when the stock runs out. */
  lastCapturer: number;
  /** Legal captures for the human's i-th hand card, in hand order. */
  takeOptions: DilotiTake[][];
  /** Declarations the human's i-th hand card could make, in hand order. */
  declareOptions: DilotiDeclCandidate[][];
  /** Whether the human's i-th hand card may simply be laid off. */
  canTrail: boolean[];
  lastResult: DilotiResult | null;
  gameEndFlag: boolean;
  winnerIdx: number;
  isHumanTurn: boolean;
  /** Recommended hand card, or -1. */
  hintHandIdx: number;
  /** "capture" | "declare" | "trail". */
  hintAction: string;
  hintTableIdxs: number[];
  hintDeclIdxs: number[];
  hintDeclValue: number;
  hintReason: string;
  config: DilotiConfig;
}
