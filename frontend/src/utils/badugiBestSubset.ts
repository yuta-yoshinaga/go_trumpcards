import type { Card } from '../types/card';

/**
 * Indices of the cards in the best Badugi subset: largest size first, ties broken by lowest rank
 * sum (lowball). Ace is encoded as `value === 1` in the project's deck and stays as 1 here.
 */
export function badugiBestSubsetIndices(cards: Card[]): number[] {
  if (cards.length === 0) return [];
  const n = cards.length;
  let bestSize = 0;
  let bestSum = Number.POSITIVE_INFINITY;
  let best: number[] = [];
  for (let mask = 1; mask < 1 << n; mask++) {
    const ranks = new Set<number>();
    const suits = new Set<string>();
    const idxs: number[] = [];
    let sum = 0;
    let ok = true;
    for (let i = 0; i < n; i++) {
      if (!(mask & (1 << i))) continue;
      const c = cards[i];
      if (ranks.has(c.value) || suits.has(c.design)) {
        ok = false;
        break;
      }
      ranks.add(c.value);
      suits.add(c.design);
      idxs.push(i);
      sum += c.value;
    }
    if (!ok) continue;
    if (idxs.length > bestSize || (idxs.length === bestSize && sum < bestSum)) {
      bestSize = idxs.length;
      bestSum = sum;
      best = idxs;
    }
  }
  return best;
}
