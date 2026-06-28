import type { Card } from '../types/card';

/** The four natural suits; Tichu specials (Dragon/Phoenix/Mahjong/Dog) use the JOKER design. */
const STANDARD_SUITS: ReadonlySet<string> = new Set(['SPADE', 'CLOVER', 'HEART', 'DIAMOND']);

/** Minimum length of a straight-flush bomb in Tichu. */
const STRAIGHT_FLUSH_MIN = 5;

/**
 * Tichu rank of a natural card. Ace (value 1) is high (14); 2..K map to 2..13.
 * Mirrors `tichuRank` in `internal/domain/TichuEval.go` (Ace high, no A-2-3-4-5 wheel).
 */
function tichuRank(card: Card): number {
  return card.value === 1 ? 14 : card.value;
}

/**
 * Returns the set of hand indices that form a **bomb** — a four-of-a-kind (four
 * cards of the same rank) or a straight flush (five or more same-suit cards with
 * consecutive ranks). Special cards (JOKER design: Dragon/Phoenix/Mahjong/Dog)
 * never participate, matching the domain's `tichuTryBomb4` / `tichuTryStraightFlush`.
 */
export function tichuBombIndices(cards: readonly Card[]): Set<number> {
  const bomb = new Set<number>();

  // Four-of-a-kind: group the natural cards by rank (a single deck holds ≤4 of each).
  const byRank = new Map<number, number[]>();
  cards.forEach((c, i) => {
    if (!STANDARD_SUITS.has(c.design)) return;
    const r = tichuRank(c);
    const arr = byRank.get(r) ?? [];
    arr.push(i);
    byRank.set(r, arr);
  });
  for (const idxs of byRank.values()) {
    if (idxs.length >= 4) for (const i of idxs) bomb.add(i);
  }

  // Straight flush: within each suit, find runs of ≥5 consecutive ranks.
  const bySuit = new Map<string, { rank: number; idx: number }[]>();
  cards.forEach((c, i) => {
    if (!STANDARD_SUITS.has(c.design)) return;
    const arr = bySuit.get(c.design) ?? [];
    arr.push({ rank: tichuRank(c), idx: i });
    bySuit.set(c.design, arr);
  });
  for (const arr of bySuit.values()) {
    arr.sort((a, b) => a.rank - b.rank);
    let run: { rank: number; idx: number }[] = [];
    const flush = () => {
      if (run.length >= STRAIGHT_FLUSH_MIN) for (const r of run) bomb.add(r.idx);
    };
    for (const entry of arr) {
      if (run.length === 0 || entry.rank === run[run.length - 1].rank + 1) {
        run.push(entry);
      } else {
        flush();
        run = [entry];
      }
    }
    flush();
  }

  return bomb;
}
