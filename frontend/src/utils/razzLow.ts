import type { Card } from '../types/card';
import { valueName } from './cardUtils';

/** The current best Razz low from a set of cards. */
export interface RazzLow {
  /** The chosen card ranks, lowest first (Ace = 1, the lowest). */
  ranks: number[];
  /** True when five distinct ranks are available (a complete low). */
  complete: boolean;
}

/**
 * Computes the best Razz low: the five lowest distinct ranks (Ace plays low).
 * Pairs don't help, so duplicate ranks are ignored; fewer than five distinct
 * ranks means the low is not yet complete.
 *
 * @param cards - The player's known cards (door + hole).
 * @returns The chosen ranks (lowest first) and whether the low is complete.
 */
export function razzBestLow(cards: Card[]): RazzLow {
  const distinct = [...new Set(cards.map((c) => c.value))].sort((a, b) => a - b);
  const ranks = distinct.slice(0, 5);
  return { ranks, complete: ranks.length === 5 };
}

/** Formats a Razz low as a high-to-low rank string, e.g. "8-6-4-3-A". */
export function formatRazzLow(low: RazzLow): string {
  return [...low.ranks]
    .sort((a, b) => b - a)
    .map((v) => valueName(v))
    .join('-');
}
