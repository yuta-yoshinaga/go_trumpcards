import type { Card } from '../types/card';

/**
 * Returns the set of indices (into the given `cards` array) that participate in
 * at least one valid Tonk meld:
 *
 * - **Set**: three or four cards of the same rank.
 * - **Run**: three or more consecutive cards of the same suit.
 *
 * This mirrors the meld shapes recognised by the Go domain's `findAllPossibleMelds`
 * (used by `FindBestMelds`). It is intentionally a "does this card belong to any
 * meld" check for UI highlighting, not an optimal hand partition.
 */
export function tonkMeldIndices(cards: Card[]): Set<number> {
  const melded = new Set<number>();

  // Sets: group indices by rank; a rank with 3+ cards forms a set.
  const byRank = new Map<number, number[]>();
  for (let i = 0; i < cards.length; i++) {
    const v = cards[i].value;
    const group = byRank.get(v);
    if (group) group.push(i);
    else byRank.set(v, [i]);
  }
  for (const group of byRank.values()) {
    if (group.length >= 3) for (const i of group) melded.add(i);
  }

  // Runs: group indices by suit, sort by rank, collect consecutive runs of 3+.
  const bySuit = new Map<string, number[]>();
  for (let i = 0; i < cards.length; i++) {
    const d = cards[i].design;
    const group = bySuit.get(d);
    if (group) group.push(i);
    else bySuit.set(d, [i]);
  }
  for (const group of bySuit.values()) {
    if (group.length < 3) continue;
    const sorted = [...group].sort((a, b) => cards[a].value - cards[b].value);
    let run: number[] = [sorted[0]];
    const flush = () => {
      if (run.length >= 3) for (const i of run) melded.add(i);
    };
    for (let k = 1; k < sorted.length; k++) {
      if (cards[sorted[k]].value === cards[run[run.length - 1]].value + 1) {
        run.push(sorted[k]);
      } else {
        flush();
        run = [sorted[k]];
      }
    }
    flush();
  }

  return melded;
}
