// Type declarations for dramaha. Split out of card.ts (issue #4366);
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

// --- Dramaha ---
//
// Dramaha is dealt and bet like Omaha, so it reuses the Hold'em wire shape —
// but the two games do NOT play the same. Two differences drive everything the
// frontend does with these types:
//
//   1. Every seat holds FIVE hole cards, not Omaha's four, and there is one
//      draw round between the flop betting and the turn where a seat may
//      exchange any number of them (`DramahaPhase.DRAW`).
//   2. The pot ALWAYS splits 50:50 between two hands made from those same five
//      cards: the Omaha hand (exactly 2 hole + exactly 3 board) and the draw
//      hand (the five hole cards as they are, board ignored).
//
// The second half is a draw hand, NOT a low hand: it is a normal
// high-poker ranking, it never fails to qualify, and it never reads the board.

/**
 * Dramaha player data (same wire shape as Hold'em).
 *
 * At showdown the backend reuses `lowBestHand` / `lowQualifies` for the DRAW
 * side of the split (`resolveShowdown` fills them from `GetDrawBestHand`, and
 * `lowQualifies` is unconditionally true because five cards always rank). They
 * carry no "8 or better" meaning here.
 */
export type DramahaPlayerData = HoldemPlayerData;

/** Dramaha CPU action (same structure as Hold'em). */
export type DramahaCpuAction = HoldemCpuAction;

/**
 * Dramaha round result (same wire shape as Hold'em).
 *
 * `hiWonAmount` is the Omaha half of the pot and `lowWonAmount` the draw half;
 * winning both is a scoop. `lowBestHand` holds the five hole cards that took
 * the draw half.
 */
export type DramahaResult = HoldemResult;

/** Dramaha side pot (same structure as Hold'em). */
export type DramahaSidePot = HoldemSidePot;

/** Dramaha equity (same structure as Hold'em). */
export type DramahaEquity = HoldemEquity;

/** Dramaha hand odds (same structure as Hold'em). */
export type DramahaHandOdds = HoldemHandOdds;

/**
 * Dramaha response (same wire shape as Hold'em).
 *
 * `phase` can additionally be `DramahaPhase.DRAW` (8). `isHiLo` stays false:
 * it means "this variant has a low hand", which Dramaha does not — the split
 * is read from the per-result `hiWonAmount` / `lowWonAmount` fields instead.
 */
export type DramahaResponse = HoldemResponse;
