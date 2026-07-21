import type { Card } from '../types/card';

/** Index sets identifying a player's qualifying low cards within their hole and the board. */
export interface LowCardIndexSets {
  loHoleSet: Set<number>;
  loBoardSet: Set<number>;
}

/**
 * Whether the community board can still yield a qualifying Omaha Hi-Lo (8-or-better) low.
 *
 * - `live`: the board already shows 3+ distinct low ranks, so a low is possible given the right hole cards.
 * - `possible`: fewer than 3 distinct low ranks so far, but enough board cards remain to reach 3.
 * - `impossible`: the board is complete (or nearly so) and can no longer reach 3 distinct low ranks.
 */
export type BoardLowStatus = 'live' | 'possible' | 'impossible';

/** Full board Hi-Lo low possibility, including the distinct low-rank count and how many more are needed. */
export interface BoardLowPossibility {
  /** Coarse possibility state derived from the current + remaining board cards. */
  status: BoardLowStatus;
  /** Distinct ranks of 8-or-lower (ace counts as low) currently on the board. */
  lowRankCount: number;
  /** Additional distinct low ranks the board still needs to make a low possible (0 once `live`). */
  needed: number;
}

/** A full Omaha/Hold'em board holds 5 community cards. */
const FULL_BOARD_SIZE = 5;
/** A low uses 3 distinct board ranks (plus 2 hole ranks) — the board must supply at least this many. */
const REQUIRED_BOARD_LOW_RANKS = 3;

/**
 * Computes whether the current community board can still make an Omaha Hi-Lo
 * (8-or-better) low possible, matching the domain rule: a low card is any rank
 * 1–8 with the ace counting low, and all ranks must be distinct. A low draws
 * exactly 3 cards from the board, so the board must eventually show at least 3
 * distinct low ranks.
 *
 * This inspects the BOARD only — it does not evaluate any player's hole cards.
 *
 * @param communityCards - The revealed community board (0–5 cards).
 * @returns The board low possibility: status, distinct low-rank count, and ranks still needed.
 */
export function boardLowPossibility(communityCards: Card[]): BoardLowPossibility {
  const lowRanks = new Set<number>();
  for (const card of communityCards) {
    if (card.value >= 1 && card.value <= 8) {
      lowRanks.add(card.value);
    }
  }
  const lowRankCount = lowRanks.size;
  const needed = Math.max(0, REQUIRED_BOARD_LOW_RANKS - lowRankCount);
  const remaining = Math.max(0, FULL_BOARD_SIZE - communityCards.length);

  let status: BoardLowStatus;
  if (lowRankCount >= REQUIRED_BOARD_LOW_RANKS) {
    status = 'live';
  } else if (needed <= remaining) {
    status = 'possible';
  } else {
    status = 'impossible';
  }
  return { status, lowRankCount, needed };
}

/** Match two cards by design and value (single deck → unique). */
function sameCard(a: Card, b: Card): boolean {
  return a.design === b.design && a.value === b.value;
}

/**
 * Maps an Omaha Hi-Lo qualifying low hand to the indices of the matching cards
 * in the player's hole and the community board, so they can be highlighted.
 *
 * @param lowBestHand - The five cards forming the qualifying low (or undefined if none).
 * @param holeCards - The player's hole cards.
 * @param communityCards - The shared board cards.
 * @returns Sets of hole and board indices that belong to the low hand.
 */
export function lowCardIndexSets(
  lowBestHand: Card[] | undefined,
  holeCards: Card[],
  communityCards: Card[],
): LowCardIndexSets {
  const loHoleSet = new Set<number>();
  const loBoardSet = new Set<number>();
  if (!lowBestHand || lowBestHand.length === 0) {
    return { loHoleSet, loBoardSet };
  }
  for (const low of lowBestHand) {
    const h = holeCards.findIndex((card) => sameCard(card, low));
    if (h >= 0) {
      loHoleSet.add(h);
      continue;
    }
    const b = communityCards.findIndex((card) => sameCard(card, low));
    if (b >= 0) loBoardSet.add(b);
  }
  return { loHoleSet, loBoardSet };
}
