// Type declarations for koikoi. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Koi-Koi phase value (0=Play 1=KoiKoiDecision 2=RoundEnd 3=GameEnd). */
export type KoiKoiPhaseValue = 0 | 1 | 2 | 3;

/** A single completed yaku (combination) with its point value. */
export interface KoiKoiYaku {
  /** Yaku identifier key (e.g. "goko", "tane", "kasu"); localized on the frontend. */
  key: string;
  /** Point value awarded for this yaku. */
  points: number;
}

/** A Koi-Koi player's public/own state. Hand `cards` are non-empty only for the human. */
export interface KoiKoiPlayer {
  id: number;
  isHuman: boolean;
  cardCount: number;
  /** Hand cards (populated only for the human). */
  cards: Card[];
  /** Cards captured into this player's pile so far this round. */
  captured: Card[];
  capturedCount: number;
  /** Cumulative match score. */
  score: number;
  /** Whether this player has called koi-koi (continue) this round. */
  calledKoiKoi: boolean;
  /** Yaku the player currently holds (from captured cards). */
  yaku: KoiKoiYaku[];
  /** Total points of the player's current yaku. */
  yakuPoints: number;
}

/** Result detail for one completed Koi-Koi round. */
export interface KoiKoiRoundResult {
  /** Winning seat index (-1 on a draw / exhausted round). */
  winner: number;
  /** Yaku the winner scored. */
  yaku: KoiKoiYaku[];
  /** Sum of yaku points before the koi-koi multiplier. */
  basePoints: number;
  /** Multiplier applied (2 when koi-koi was called). */
  multiplier: number;
  /** Final points awarded (basePoints × multiplier). */
  total: number;
  /** Number of koi-koi calls that occurred in the round. */
  koikoiCount: number;
}

/** Koi-Koi game configuration. */
export interface KoiKoiConfig {
  cpuDifficulty: number;
  targetScore: number;
}

/** A suggested hint for Koi-Koi, computed by the backend. */
export interface KoiKoiHint {
  cardIndex: number;
  fieldIndex: number;
  /** 1 = suggest calling koi-koi, 0 = suggest stopping (shobu). */
  koikoi: number;
  /** i18n reason suffix identifier. */
  reason: string;
}

/**
 * Full Koi-Koi (こいこい) game state returned from the API.
 *
 * Koi-Koi is a 2-player hanafuda capture game. On the human's turn the player
 * plays a hand card to capture a matching field card (same month), then draws
 * from the stock which may also capture. When a new yaku (combination) is
 * completed the player decides between koi-koi (continue for more, doubling the
 * stakes) and shobu (stop and score). Cards carry a hanafuda face descriptor
 * (`glyph`/`label`/`color`/`deck`) and render procedurally via `CardImage`.
 */
export interface KoiKoiResponse extends BaseGameResponse {
  players: KoiKoiPlayer[];
  phase: KoiKoiPhaseValue;
  /** Round (deal) counter. */
  roundNumber: number;
  /** Seat index of the player whose turn it is to act. */
  currentTurn: number;
  /** Cards currently face-up on the field. */
  fieldCards: Card[];
  /** Number of cards left in the stock. */
  remainingDeck: number;
  /** Number of koi-koi calls so far this round (drives the multiplier). */
  koikoiCount: number;
  /** Indices in the human's hand that are legal to play (human Play turn). */
  playableIndices: number[];
  /** Map of hand index -> field indices that hand card can capture (present when 2-way choice). */
  captureOptions: Record<number, number[]>;
  /** Yaku the acting player just completed, pending a koi-koi/shobu decision. */
  pendingYaku: KoiKoiYaku[];
  /** Total points of the pending yaku. */
  pendingPoints: number;
  /** Winning seat index of the just-finished round, or -1. */
  roundWinner: number;
  /** Winning seat index of the whole match, or -1. */
  winner: number;
  /** Whether the match has ended. */
  gameEndFlag: boolean;
  /** Whether it is currently the human's turn to act. */
  isHumanTurn: boolean;
  /** Result of the most recently completed round (RoundEnd phase), or null. */
  lastRoundResult?: KoiKoiRoundResult | null;
  hint?: KoiKoiHint | null;
  config: KoiKoiConfig;
}

// --- Hachi-Hachi (八八 / はちはち) ---
