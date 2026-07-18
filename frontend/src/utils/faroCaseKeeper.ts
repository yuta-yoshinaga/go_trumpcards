import type { Card } from '../types/card';

/** Number of cards of each rank in the standard 52-card Faro deck. */
export const FARO_RANK_COUNT = 4;

/** Rank values tracked by the case keeper, A (1) through K (13). */
export const FARO_RANKS: readonly number[] = [1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13];

/**
 * Returns a stable identity key for a card (suit design + value). Every card in
 * a standard 52-card deck maps to a unique key, so a Set of keys deduplicates
 * cards that are revealed again by an idempotent state refresh.
 */
export function faroCardKey(card: Card): string {
  return `${card.design}-${card.value}`;
}

/**
 * Merges newly revealed cards into the running set of seen-card keys. Null or
 * undefined entries are ignored, and duplicates are naturally deduped by the
 * Set. Returns a new Set (the input is never mutated).
 */
export function mergeSeenCards(prev: ReadonlySet<string>, cards: ReadonlyArray<Card | null | undefined>): Set<string> {
  const next = new Set(prev);
  for (const card of cards) {
    if (card) next.add(faroCardKey(card));
  }
  return next;
}

/**
 * Computes how many cards of each rank (1..13) remain unseen, given the set of
 * already-seen card keys. Each rank starts at {@link FARO_RANK_COUNT}; every
 * distinct seen card of that rank subtracts one. Values are clamped to
 * `[0, FARO_RANK_COUNT]`, and keys with out-of-range ranks are ignored.
 */
export function remainingByRank(seen: ReadonlySet<string>): Record<number, number> {
  const remaining: Record<number, number> = {};
  for (const rank of FARO_RANKS) remaining[rank] = FARO_RANK_COUNT;
  for (const key of seen) {
    const value = Number(key.slice(key.lastIndexOf('-') + 1));
    if (Number.isInteger(value) && remaining[value] !== undefined && remaining[value] > 0) {
      remaining[value] -= 1;
    }
  }
  return remaining;
}
