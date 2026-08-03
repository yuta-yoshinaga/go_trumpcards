// Type declarations for ganjifa. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Ganjifa game phase (0=Play 1=TrickEnd 2=RoundEnd 3=GameEnd). */
export type GanjifaPhaseValue = 0 | 1 | 2 | 3;

/** A Ganjifa player's public/own state. Cards are non-empty only for the human. */
export interface GanjifaPlayer {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  trickCount: number;
  /** Cumulative match score of this individual player (one point per trick). */
  score: number;
}

/** A card played into the current Ganjifa trick. */
export interface GanjifaTrickCard {
  playerIdx: number;
  card: Card;
}

/** Ganjifa game configuration. */
export interface GanjifaConfig {
  cpuDifficulty: number;
  targetRounds: number;
}

/** A suggested hint for Ganjifa, computed by the backend. */
export interface GanjifaHint {
  cardIndices: number[];
  /** i18n reason suffix identifier. */
  reason: string;
}

/** Full Ganjifa game state returned from the API. */
export interface GanjifaResponse extends BaseGameResponse {
  players: GanjifaPlayer[];
  phase: GanjifaPhaseValue;
  roundNumber: number;
  trickNumber: number;
  currentPlayerIdx: number;
  leadPlayerIdx: number;
  dealerIdx: number;
  /**
   * Trump suit as a Ganjifa design (1-8), auto-declared from the dealer's
   * longest suit. 1-4 are the strong group (12 high), 5-8 the weak group
   * (1 high) — see {@link isGanjifaStrongSuit}.
   */
  trumpSuit: number;
  currentTrick: GanjifaTrickCard[];
  /** Cumulative match scores per player — [p0, p1, p2]. */
  playerScores: number[];
  /** Tricks captured per player this round — [p0, p1, p2]. */
  roundTricks: number[];
  /** Indices in the human's hand that are legal to play (non-empty on human Play turn). */
  playableIndices: number[];
  gameEndFlag: boolean;
  /** Winning player seat index, or -1 until the game ends (also -1 on a tie). */
  winnerPlayer: number;
  /** Whether it is currently the human's turn to play a card. */
  isHumanTurn: boolean;
  hint?: GanjifaHint | null;
  config: GanjifaConfig;
}

/** Number of Ganjifa suits (designs 1-8). */
export const GANJIFA_SUIT_COUNT = 8;

/** Highest design in the strong suit group. Mirrors `domain.GanjifaStrongSuitMax`. */
export const GANJIFA_STRONG_SUIT_MAX = 4;

/**
 * Whether a Ganjifa design belongs to the strong suit group.
 *
 * This is the one rule that separates Ganjifa from every other trick-taker
 * here: in a strong suit (1-4) the 12 is the highest card, but in a weak suit
 * (5-8) the **1** is, so comparing raw rank values gives the opposite answer
 * across half the deck.
 */
export function isGanjifaStrongSuit(design: number): boolean {
  return design >= 1 && design <= GANJIFA_STRONG_SUIT_MAX;
}

/** Display glyphs for the 8 Ganjifa suits, indexed by design (index 0 unused). */
export const GANJIFA_SUIT_GLYPHS = ['', '♛', '†', '◉', '♟', '♪', '✦', '▤', '▧'] as const;

/** Traditional Mughal names for the 8 Ganjifa suits, indexed by design (index 0 unused). */
export const GANJIFA_SUIT_NAMES = [
  '',
  'Taj',
  'Shamsher',
  'Ashrafi',
  'Ghulam',
  'Chang',
  'Surkh',
  'Barat',
  'Qimash',
] as const;

/** Formats a Ganjifa design as "<glyph> <name>", or "?" when out of range. */
export function formatGanjifaSuit(design: number): string {
  if (design < 1 || design > GANJIFA_SUIT_COUNT) return '?';
  return `${GANJIFA_SUIT_GLYPHS[design]} ${GANJIFA_SUIT_NAMES[design]}`;
}
