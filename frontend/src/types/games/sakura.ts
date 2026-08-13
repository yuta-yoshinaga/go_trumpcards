// Type declarations for sakura. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Sakura phase value (0=Play 1=RoundEnd 2=GameEnd). */
export type SakuraPhaseValue = 0 | 1 | 2;

/** A single bonus combination with its point value. */
export interface SakuraBonus {
  /** Bonus identifier key ("sakuraSake" | "allBrights"); localized on the frontend. */
  key: string;
  /** Point value awarded for this bonus. */
  points: number;
}

/** A Sakura player's public/own state. Hand `cards` are non-empty only for the human. */
export interface SakuraPlayer {
  id: number;
  /** Seat display name from the server. */
  name: string;
  isHuman: boolean;
  cardCount: number;
  /** Hand cards (populated only for the human). */
  cards: Card[];
  /** Cards captured into this player's pile so far this round. */
  taken: Card[];
  takenCount: number;
  /** Points from captured cards alone (20/10/5/1 per card). */
  cardPoints: number;
  /** Bonus combinations currently held. */
  bonuses: SakuraBonus[];
  /** Total points from bonuses. */
  bonusPoints: number;
  /** cardPoints + bonusPoints. */
  totalPoints: number;
  /** Cumulative match score across rounds. */
  score: number;
  /** Points scored in the round just finished. */
  roundScore: number;
  /** Number of rounds this seat has won. */
  roundWins: number;
}

/** One seat's breakdown in a completed round. */
export interface SakuraSeatResult {
  cardPoints: number;
  bonuses: SakuraBonus[];
  bonusPoints: number;
  total: number;
}

/** Result detail for one completed Sakura round. */
export interface SakuraRoundResult {
  /** Round number this result belongs to. */
  round: number;
  /** Winning seat index (-1 when the round is tied). */
  winner: number;
  /** Per-seat breakdown, in seat order. */
  seats: SakuraSeatResult[];
}

/** Sakura game configuration. */
export interface SakuraConfig {
  seats: number;
  rounds: number;
}

/** A suggested hint for Sakura, computed by the backend. */
export interface SakuraHint {
  cardIndex: number;
  /** Field card to take, or -1 when nothing matches. */
  fieldIndex: number;
  /** i18n reason suffix identifier ("capture" | "discard"). */
  reason: string;
}

/**
 * Full Sakura (さくら / 肥後花) game state returned from the API.
 *
 * Sakura is a Kumamoto hanafuda matching game for 2-4 players. **It is scored by
 * adding up captured cards (20/10/5/1 points), not by completing yaku** — which
 * is what separates it from Koi-Koi and Hachi-Hachi. On a turn the player plays
 * one hand card to take a field card of the same month, then flips one stock
 * card that resolves the same way. Cards carry a hanafuda face descriptor
 * (`glyph`/`label`/`color`/`deck`) and render procedurally via `CardImage`.
 */
export interface SakuraResponse extends BaseGameResponse {
  players: SakuraPlayer[];
  phase: SakuraPhaseValue;
  /** Current round number (1-based). */
  round: number;
  /** Rounds configured for the match. */
  totalRounds: number;
  /** Seat index of the player whose turn it is to act. */
  currentTurn: number;
  /** Seat index of the dealer, who leads the round. */
  dealer: number;
  /** Cards currently face-up on the field. */
  fieldCards: Card[];
  /** Number of cards left in the stock. */
  stockCount: number;
  /** Map of hand index -> field indices that hand card can take. */
  captureOptions: Record<number, number[]>;
  /** Winning seat index of the whole match, or -1. */
  winner: number;
  /** Whether the match has ended. */
  gameEndFlag: boolean;
  /** Whether it is currently the human's turn to act. */
  isHumanTurn: boolean;
  /** Result of the most recently completed round (RoundEnd phase), or null. */
  lastResult?: SakuraRoundResult | null;
  hint?: SakuraHint | null;
  config: SakuraConfig;
}
