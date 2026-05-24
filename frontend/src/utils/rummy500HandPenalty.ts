import type { Card } from '../types/card';

/** Per-card penalty value used when a round ends with cards still in hand. A=1, J/Q/K=10, otherwise face value. */
export function rummy500CardPenalty(value: number): number {
  if (value === 1) return 1;
  if (value >= 10) return 10;
  return value;
}

/** Total penalty (sum of card values) for a Rummy 500 hand. */
export function rummy500HandPenalty(cards: Card[]): number {
  let total = 0;
  for (const c of cards) {
    total += rummy500CardPenalty(c.value);
  }
  return total;
}
