// Type declarations for omaha. Split out of card.ts (issue #4366);
// card.ts re-exports this file, so existing imports keep working.

import type {
  HoldemCpuAction,
  HoldemEquity,
  HoldemHandOdds,
  HoldemPlayerData,
  HoldemResponse,
  HoldemResult,
  HoldemSidePot,
} from './holdem';

// --- Omaha Hold'em ---
// Omaha shares identical response/player structures with Holdem
/** Omaha player data (same structure as Hold'em). */
export type OmahaPlayerData = HoldemPlayerData;

/** Omaha CPU action (same structure as Hold'em). */
export type OmahaCpuAction = HoldemCpuAction;

/** Omaha round result (same structure as Hold'em). */
export type OmahaResult = HoldemResult;

/** Omaha side pot (same structure as Hold'em). */
export type OmahaSidePot = HoldemSidePot;

/** Omaha equity (same structure as Hold'em). */
export type OmahaEquity = HoldemEquity;

/** Omaha hand odds (same structure as Hold'em). */
export type OmahaHandOdds = HoldemHandOdds;

/** Omaha response (same structure as Hold'em). */
export type OmahaResponse = HoldemResponse;

// --- Short Deck Hold'em ---
