import type { Card } from '../types/card';

/** Backend's GinRummyKnockThreshold (10). */
export const GIN_RUMMY_KNOCK_THRESHOLD = 10;

/** Gin Rummy card value (A=1, 2-9=face, 10/J/Q/K=10). */
export function ginRummyCardValue(card: Card): number {
  if (card.value === 1) return 1;
  if (card.value >= 10) return 10;
  return card.value;
}

/** Sum of card values in a deadwood pile. */
export function calcDeadwoodValue(cards: readonly Card[]): number {
  let total = 0;
  for (const c of cards) total += ginRummyCardValue(c);
  return total;
}

/** Indexed handle so the meld-finder can compare card identity rather than
 * value (suits matter for runs and identical-rank duplicates would otherwise
 * collide). */
interface IndexedCard {
  idx: number;
  card: Card;
}

/** Find the minimum-deadwood split of `hand` into melds + deadwood.
 * Mirrors `FindBestMelds` in internal/domain/GinRummy.go but the only
 * value we surface back to React is the integer deadwood total. */
export function bestDeadwoodValue(hand: readonly Card[]): number {
  if (hand.length === 0) return 0;
  const indexed: IndexedCard[] = hand.map((card, idx) => ({ idx, card }));
  return search(indexed);
}

function search(remaining: IndexedCard[]): number {
  const candidates = enumerateMelds(remaining);
  let best = calcDeadwoodValue(remaining.map((r) => r.card));
  if (candidates.length === 0) return best;
  for (const meld of candidates) {
    const meldIdx = new Set(meld.map((m) => m.idx));
    const rest = remaining.filter((r) => !meldIdx.has(r.idx));
    const dv = search(rest);
    if (dv < best) best = dv;
    if (best === 0) break;
  }
  return best;
}

function enumerateMelds(cards: IndexedCard[]): IndexedCard[][] {
  const out: IndexedCard[][] = [];

  // Sets: 3+ same value across any suit
  const byValue = new Map<number, IndexedCard[]>();
  for (const c of cards) {
    const list = byValue.get(c.card.value) ?? [];
    list.push(c);
    byValue.set(c.card.value, list);
  }
  for (const group of byValue.values()) {
    if (group.length >= 3) {
      out.push(group.slice(0, 3));
      if (group.length >= 4) out.push(group.slice(0, 4));
    }
  }

  // Runs: 3+ consecutive in the same suit
  const bySuit = new Map<string, IndexedCard[]>();
  for (const c of cards) {
    const list = bySuit.get(c.card.design) ?? [];
    list.push(c);
    bySuit.set(c.card.design, list);
  }
  for (const group of bySuit.values()) {
    if (group.length < 3) continue;
    const sorted = [...group].sort((a, b) => a.card.value - b.card.value);
    for (let i = 0; i < sorted.length; i++) {
      const run: IndexedCard[] = [sorted[i]];
      for (let j = i + 1; j < sorted.length; j++) {
        if (sorted[j].card.value === run[run.length - 1].card.value + 1) {
          run.push(sorted[j]);
          if (run.length >= 3) out.push([...run]);
        } else {
          break;
        }
      }
    }
  }
  return out;
}
