// Type declarations for bauernschnapsen. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Bauernschnapsen player data with team, trick count, and hand. */
export interface BauernschnapsenPlayerData {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  team: number;
  trickCount: number;
}

/** A card played in a Bauernschnapsen trick. */
export interface BauernschnapsenTrickCard {
  playerIdx: number;
  card: Card;
}

/** Bauernschnapsen game configuration. */
export interface BauernschnapsenConfig {
  cpuDifficulty: number;
  targetScore: number;
}

/** A suggested hint for Bauernschnapsen. */
export interface BauernschnapsenHint {
  cardIndex?: number;
  reason: string;
  isMarriage: boolean;
}

/** Full Bauernschnapsen game state returned from the API. */
export interface BauernschnapsenResponse extends BaseGameResponse {
  players: BauernschnapsenPlayerData[];
  phase: number;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  dealerIdx: number;
  trumpSuit: number;
  /** 採用された契約 (0=未宣言 1=通常 2=同スート縛り 3=ベテル)。 */
  contract: number;
  /** 契約を宣言した席。-1 は未確定 (契約フェーズ中)。 */
  declarerIdx: number;
  currentTrick: BauernschnapsenTrickCard[];
  teamScores: number[];
  roundPoints: number[];
  roundMarriage: number[];
  /** 出せる手札位置。**追従はトリック 1 から必須**なので、これに無い札は押せない。 */
  validPlayIndices: number[];
  marriageIndices: number[];
  gameEndFlag: boolean;
  winnerTeam: number;
  leadPlayerIdx: number;
  config: BauernschnapsenConfig;
  hint?: BauernschnapsenHint;
}

// --- Contract Bridge (コントラクトブリッジ) ---
