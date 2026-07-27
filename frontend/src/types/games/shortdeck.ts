// Type declarations for shortdeck. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type { HoldemEquity, HoldemHandOdds, HoldemPlayerData, HoldemResponse, HoldemSidePot } from './holdem';

/** Short Deck Hold'em player data (same structure as Hold'em). */
export type ShortDeckPlayerData = HoldemPlayerData;

/** Short Deck Hold'em side pot (same structure as Hold'em). */
export type ShortDeckSidePot = HoldemSidePot;

/** Short Deck Hold'em equity (same structure as Hold'em). */
export type ShortDeckEquity = HoldemEquity;

/** Short Deck Hold'em hand odds (same structure as Hold'em). */
export type ShortDeckHandOdds = HoldemHandOdds;

/** Short Deck Hold'em response (same structure as Hold'em). */
export type ShortDeckResponse = HoldemResponse;

// --- Hearts ---
