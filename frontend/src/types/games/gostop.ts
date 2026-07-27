// Type declarations for gostop. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Go-Stop phase value (0=Play 1=GoDecision 2=RoundEnd 3=GameEnd). */
export type GoStopPhaseValue = 0 | 1 | 2 | 3;

/**
 * Korean Go-Stop scoring breakdown for a player's captured pile. Points are
 * split into the five categories (gwang/godori/tti/yeol/pi), then the `base`
 * total is multiplied by the go multiplier to produce `goScore`.
 */
export interface GoStopBreakdown {
  /** Bright (光 / 광) points. */
  gwang: number;
  /** Five-bird (五鳥 / 고도리) points. */
  godori: number;
  /** Ribbon (띠) points. */
  tti: number;
  /** Animal (열끗) points. */
  yeol: number;
  /** Junk (피) points. */
  pi: number;
  /** Sum of all category points before the go multiplier. */
  base: number;
  /** Number of "go" calls made this round. */
  goCount: number;
  /** Multiplier applied for the go calls. */
  goMult: number;
  /** Points after applying the go multiplier. */
  goScore: number;
  /** Number of bright cards captured. */
  brightCount: number;
  /** Number of ribbon cards captured. */
  ribbonCount: number;
  /** Number of animal cards captured. */
  animalCount: number;
  /** Number of junk cards captured. */
  piCount: number;
}

/** A Go-Stop player's public/own state. Hand `cards` are non-empty only for the human. */
export interface GoStopPlayer {
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
  /** Number of "go" calls this player has made this round. */
  goCount: number;
  /** Current scoring breakdown from captured cards, or null. */
  breakdown: GoStopBreakdown | null;
  /** Current total points (base × go multiplier). */
  points: number;
}

/** Result detail for one completed Go-Stop round. */
export interface GoStopRoundResult {
  /** Winning seat index (-1 on a draw / exhausted round). */
  winner: number;
  /** The winner's scoring breakdown. */
  breakdown: GoStopBreakdown | null;
  /** Base points before go/bak multipliers. */
  basePoints: number;
  /** Points after the go multiplier. */
  goScore: number;
  /** Combined bak (penalty-doubling) multiplier applied to the loser's payment. */
  bakMult: number;
  /** Final points transferred to the winner. */
  total: number;
  /** Whether gwang-bak (bright penalty) applied. */
  gwangBak: boolean;
  /** Whether pi-bak (junk penalty) applied. */
  piBak: boolean;
  /** Whether go-bak (go-call penalty) applied. */
  goBak: boolean;
  /** Number of go calls in the round. */
  goCount: number;
}

/** Go-Stop game configuration. */
export interface GoStopConfig {
  cpuDifficulty: number;
  targetScore: number;
}

/** A suggested hint for Go-Stop, computed by the backend. */
export interface GoStopHint {
  cardIndex: number;
  fieldIndex: number;
  /** 1 = suggest calling go, 0 = suggest stopping (-1 during Play). */
  go: number;
  /** i18n reason suffix identifier. */
  reason: string;
}

/**
 * Full Go-Stop (Godori / ゴーストップ) game state returned from the API.
 *
 * Go-Stop is the Korean sibling of Koi-Koi, played with the same 48-card
 * hanafuda deck. On the human's turn the player plays a hand card to capture a
 * matching field card (same month), then draws from the stock which may also
 * capture. When the target score is reached the GoDecision phase offers go
 * (continue for more) or stop (bank the points). Cards carry a hanafuda face
 * descriptor (`glyph`/`label`/`color`/`deck`) and render procedurally via
 * `CardImage`.
 */
export interface GoStopResponse extends BaseGameResponse {
  players: GoStopPlayer[];
  phase: GoStopPhaseValue;
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
  /** Breakdown of the score that triggered the go/stop decision, or null. */
  pendingBreakdown: GoStopBreakdown | null;
  /** Total points pending a go/stop decision. */
  pendingPoints: number;
  /** Winning seat index of the just-finished round, or -1. */
  roundWinner: number;
  /** Winning seat index of the whole match, or -1. */
  winner: number;
  /** Raw result enum value. */
  result: number;
  /** Whether the match has ended. */
  gameEndFlag: boolean;
  /** Whether it is currently the human's turn to act. */
  isHumanTurn: boolean;
  /** Result of the most recently completed round (RoundEnd phase), or null. */
  lastRoundResult?: GoStopRoundResult | null;
  hint?: GoStopHint | null;
  config: GoStopConfig;
}

// --- Tablanet (Tablić) ---
