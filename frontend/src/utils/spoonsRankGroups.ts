import type { Card } from '../types/card';

/**
 * Per-card grouping info for a Spoons hand, computed purely on the frontend by
 * same-rank (`value`) equality. Cards whose rank appears 2+ times in the hand
 * form a group and share a `colorIndex`; singleton ranks get `colorIndex: null`.
 */
export interface SpoonsRankGroup {
  /**
   * Deterministic group color index (0-based) for ranks with 2+ cards, or
   * `null` for singleton ranks. Assigned by ascending rank so the mapping is
   * stable regardless of card order in the hand.
   */
  colorIndex: number | null;
  /** Number of cards in the hand sharing this card's rank (always >= 1). */
  count: number;
}

/**
 * Computes same-rank group info for each card in a Spoons hand, returned as an
 * array parallel to `hand`. Ranks appearing 2+ times are assigned a stable
 * `colorIndex` (by ascending rank); singletons get `null`. This lets the UI
 * color-code cards so the player can see how close they are to four-of-a-kind.
 */
export function computeSpoonsRankGroups(hand: Card[]): SpoonsRankGroup[] {
  const counts = new Map<number, number>();
  for (const card of hand) {
    counts.set(card.value, (counts.get(card.value) ?? 0) + 1);
  }

  // Ranks with 2+ cards get a stable color index, assigned by ascending rank.
  const groupedRanks = [...counts.entries()]
    .filter(([, count]) => count >= 2)
    .map(([rank]) => rank)
    .sort((a, b) => a - b);
  const colorIndexByRank = new Map<number, number>();
  for (const [i, rank] of groupedRanks.entries()) {
    colorIndexByRank.set(rank, i);
  }

  return hand.map((card) => ({
    colorIndex: colorIndexByRank.has(card.value) ? (colorIndexByRank.get(card.value) as number) : null,
    count: counts.get(card.value) ?? 1,
  }));
}
