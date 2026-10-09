import type { Card } from '../types/card';
import { valueName } from './cardUtils';

/** The current best Razz low from a set of cards. */
export interface RazzLow {
  /** The chosen card ranks, lowest first (Ace = 1, the lowest). */
  ranks: number[];
  /** True when five distinct ranks are available (a complete low). */
  complete: boolean;
  /** Indices of the selected cards in the input array, in ascending rank order. */
  cardIndices: number[];
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
  const selected = new Map<number, number>();
  cards.forEach((card, index) => {
    if (!selected.has(card.value)) selected.set(card.value, index);
  });
  const chosen = [...selected.entries()].sort(([a], [b]) => a - b).slice(0, 5);
  const ranks = chosen.map(([rank]) => rank);
  return { ranks, complete: ranks.length === 5, cardIndices: chosen.map(([, index]) => index) };
}

/** Formats a Razz low as a high-to-low rank string, e.g. "8-6-4-3-A". */
export function formatRazzLow(low: RazzLow): string {
  return [...low.ranks]
    .sort((a, b) => b - a)
    .map((v) => valueName(v))
    .join('-');
}
