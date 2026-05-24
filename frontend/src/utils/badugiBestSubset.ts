import type { Card } from '../types/card';

/** Numeric Badugi rank used for lowball comparison (A=1, then 2..K). */
function badugiRank(value: number): number {
  if (value === 1) return 1;
  return value;
}

/** Indices of the cards in the best Badugi subset (largest size, then lowest sum). */
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
      const r = badugiRank(c.value);
      if (ranks.has(r) || suits.has(c.design)) {
        ok = false;
        break;
      }
      ranks.add(r);
      suits.add(c.design);
      idxs.push(i);
      sum += r;
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
