// Type declarations for hachihachi. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Hachi-Hachi phase value (0=Play 1=RoundEnd 2=GameEnd). */
export type HachiHachiPhaseValue = 0 | 1 | 2;

/** A single completed yaku (combination) with its bonus point value. */
export interface HachiHachiYaku {
  /** Yaku identifier key (e.g. "goko", "inoshikacho"); localized on the frontend. */
  key: string;
  /** Bonus point value awarded for this yaku. */
  points: number;
}

/** A Hachi-Hachi player's public/own state. Hand `cards` are non-empty only for the human. */
export interface HachiHachiPlayer {
  id: number;
  isHuman: boolean;
  cardCount: number;
  /** Hand cards (populated only for the human). */
  cards: Card[];
  /** Cards captured into this player's pile so far this round. */
  captured: Card[];
  capturedCount: number;
  /** Cumulative signed match score (each round settled against the 88 baseline). */
  score: number;
  /** This player's signed delta from the most recently completed round. */
  roundDelta: number;
  /** Raw card-point total of the captured pile (Bright 20 / Animal 10 / Ribbon 5 / Chaff 1). */
  rawScore: number;
  /** Yaku the player currently holds (from captured cards). */
  yaku: HachiHachiYaku[];
}

/** One player's round-settlement breakdown, present in {@link HachiHachiRoundResult}. */
export interface HachiHachiPlayerScore {
  playerIdx: number;
  /** Raw card-point total for the round. */
  rawScore: number;
  /** Yaku bonuses the player scored this round. */
  yaku: HachiHachiYaku[];
  /** Sum of yaku bonus points. */
  bonus: number;
  /** Signed delta from the 88 baseline (rawScore + bonus − 88). */
  delta: number;
}

/** Result detail for one completed Hachi-Hachi round (RoundEnd phase). */
export interface HachiHachiRoundResult {
  /** Per-player settlement breakdowns for the round. */
  scores: HachiHachiPlayerScore[];
  /** Seat index with the highest delta this round. */
  best: number;
}

/** Hachi-Hachi game configuration. */
export interface HachiHachiConfig {
  cpuDifficulty: number;
  /** Number of rounds (deals) played before the match is settled. */
  targetRounds: number;
}

/** A suggested hint for Hachi-Hachi, computed by the backend. */
export interface HachiHachiHint {
  cardIndex: number;
  fieldIndex: number;
  /** i18n reason suffix identifier (e.g. "capture", "discard_low"). */
  reason: string;
}

/**
 * Full Hachi-Hachi (八八) game state returned from the API.
 *
 * Hachi-Hachi is the classic 3-player Japanese hanafuda game on the 48-card
 * flower deck. Players capture field cards of the same month with their hand
 * and stock draws; when every hand is exhausted the round's captured piles are
 * scored by card points (Bright 20 / Animal 10 / Ribbon 5 / Chaff 1) plus yaku
 * bonuses, and each player settles against the 88 baseline. Unlike Koi-Koi
 * there is no koi-koi/stop decision — phases are simply Play, RoundEnd, and
 * GameEnd. Cards carry a hanafuda face descriptor (`glyph`/`label`/`color`/
 * `deck`) and render procedurally via `CardImage`.
 */
export interface HachiHachiResponse extends BaseGameResponse {
  players: HachiHachiPlayer[];
  phase: HachiHachiPhaseValue;
  /** Round (deal) counter. */
  roundNumber: number;
  /** Seat index of the player whose turn it is to act. */
  currentTurn: number;
  /** Cards currently face-up on the field. */
  fieldCards: Card[];
  /** Number of cards left in the stock. */
  remainingDeck: number;
  /** Indices in the human's hand that are legal to play (human Play turn). */
  playableIndices: number[];
  /** Map of hand index -> field indices that hand card can capture (present when 2-way choice). */
  captureOptions: Record<number, number[]>;
  /** Winning seat index of the whole match (highest cumulative score), or -1. */
  winner: number;
  /** Match result enum value from the backend. */
  result: number;
  /** Whether the match has ended. */
  gameEndFlag: boolean;
  /** Whether it is currently the human's turn to act. */
  isHumanTurn: boolean;
  /** Settlement of the most recently completed round (RoundEnd phase), or null. */
  lastRoundResult?: HachiHachiRoundResult | null;
  hint?: HachiHachiHint | null;
  config: HachiHachiConfig;
}

// --- Go-Stop (Godori / ゴーストップ) ---
