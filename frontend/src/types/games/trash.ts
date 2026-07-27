// Type declarations for trash. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { BaseGameResponse, Card } from '../common';

/** A single slot (position 1..10) for one Trash player. */
export interface TrashSlot {
  /** Face-down slots omit this field. Only face-up cards expose their identity. */
  card?: Card;
  faceUp: boolean;
}

/** One player's full Trash state: 10 slots plus a flag for the CPU. */
export interface TrashPlayerState {
  slots: TrashSlot[];
  isCpu: boolean;
}

/** API response shape for a Trash game. */
export interface TrashResponse extends BaseGameResponse {
  phase: number;
  current: number;
  players: [TrashPlayerState, TrashPlayerState];
  stockSize: number;
  discardSize: number;
  discardTop?: Card;
  pending?: Card;
  moveCount: number;
  winner: number;
}

// --- Whist (ホイスト) ---
