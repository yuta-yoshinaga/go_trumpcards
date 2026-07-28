// Type declarations for loo. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** Loo game phase (0=Decide 1=Play 2=TrickEnd 3=RoundEnd). */
export type LooPhaseValue = 0 | 1 | 2 | 3;

/** A Loo player's public/own state. Cards are non-empty only for the human. */
export interface LooPlayer {
  id: number;
  isHuman: boolean;
  cardCount: number;
  cards: Card[];
  trickCount: number;
  /** Whether this player is participating in the current deal (play vs pass). */
  playing: boolean;
  /** Cumulative chip balance of this individual player (can be negative). */
  chips: number;
}

/** A card played into the current Loo trick. */
export interface LooTrickCard {
  playerIdx: number;
  card: Card;
}

/** Per-deal settlement breakdown for Loo (surfaced at round end). */
export interface LooDealDetail {
  /** Pot size at the start of the deal (used to size the per-trick payout). */
  potStart: number;
  /** Trump suit for the scored deal (1=♠ 2=♣ 3=♥ 4=♦). */
  trumpSuit: number;
  /** Whether each seat participated (played) this deal, keyed by seat index. */
  playing: boolean[];
  /** Tricks captured per participating seat this deal, keyed by seat index. */
  tricks: Record<number, number>;
  /** Chips gained (or lost) per player this deal, keyed by seat index. */
  gained: Record<number, number>;
  /** Seat indices of players who were "looed" (played but took no tricks). */
  looed: number[];
  /** Chips carried over into the next deal's pot. */
  potCarry: number;
}

/** Loo game configuration. */
export interface LooConfig {
  cpuDifficulty: number;
  ante: number;
}

/** A suggested hint for Loo, computed by the backend. */
export interface LooHint {
  cardIndices: number[];
  /** Suggested participation decision (present for a decide-phase hint). */
  decision?: boolean | null;
  /** i18n reason suffix identifier. */
  reason: string;
}

/**
 * Full Loo (Lanterloo) game state returned from the API.
 *
 * Loo is a 4-player (1 human + 3 CPU, individual chips) pot-based gambling
 * trick-taker on a standard 52-card deck. Each player is dealt five cards; the
 * turn-up card sets trump. Players decide to play or pass, then participants
 * fight five must-follow / must-head tricks, each trick winning one-fifth of the
 * pot. A participant who wins no trick is "looed" and pays a penalty into the
 * next deal's pot. There is no game-over target — it is a repeating deal loop, so
 * `gameEndFlag` is always false.
 */
export interface LooResponse extends BaseGameResponse {
  players: LooPlayer[];
  phase: LooPhaseValue;
  roundNumber: number;
  trickNumber: number;
  totalTricks: number;
  dealerIdx: number;
  /** Seat index of the player whose turn it is to act. */
  currentTurn: number;
  /** Seat index of the player whose turn it is to decide (play/pass). */
  decidePlayerIdx: number;
  /** Trump suit (0=unset, 1=♠ 2=♣ 3=♥ 4=♦). */
  trumpSuit: number;
  /** The turn-up card whose suit becomes trump. */
  turnUp?: Card | null;
  /** Current pot size (chips available for distribution this deal). */
  pot: number;
  /** Pot size at the start of the deal (used to size the per-trick payout). */
  potStart: number;
  currentTrick: LooTrickCard[];
  lastTrick: LooTrickCard[];
  /** Seat index of the last trick winner, or -1. */
  lastTrickWinner: number;
  /** Indices in the human's hand that are legal to play (non-empty on human Play turn). */
  playableIndices: number[];
  /** Always false — Loo has no game-over; it is a repeating deal loop. */
  gameEndFlag: boolean;
  lastDealDetail?: LooDealDetail | null;
  /** Whether it is currently the human's turn to act. */
  isHumanTurn: boolean;
  hint?: LooHint | null;
  config: LooConfig;
}

// --- Basra (Bastra) ---
