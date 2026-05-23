import type { Card, CardDesign } from '../types/card';

/**
 * Returns the suit (design) that forms the longest flush in a High Card Flush
 * hand, or null when the hand is empty. Tie-breaker: when two suits have the
 * same count, prefers the one containing the highest single card — matching
 * High Card Flush's own "high card wins" tiebreaker so the highlighted set
 * matches the suit the dealer would compare against.
 */
export function longestFlushSuit(hand: readonly Card[]): CardDesign | null {
  if (hand.length === 0) return null;
  const counts = new Map<CardDesign, { count: number; highest: number }>();
  for (const c of hand) {
    const cur = counts.get(c.design);
    if (cur === undefined) {
      counts.set(c.design, { count: 1, highest: c.value });
    } else {
      cur.count += 1;
      if (c.value > cur.highest) cur.highest = c.value;
    }
  }
  let best: { suit: CardDesign; count: number; highest: number } | null = null;
  for (const [suit, info] of counts) {
    if (best === null || info.count > best.count || (info.count === best.count && info.highest > best.highest)) {
      best = { suit, count: info.count, highest: info.highest };
    }
  }
  return best?.suit ?? null;
}
