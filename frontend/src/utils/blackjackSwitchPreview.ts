import type { Card } from '../types/card';

/** Numeric value contributed by one card in Blackjack scoring. Aces report 11; soft-total adjustment happens in {@link blackjackScore}. */
function bjCardBase(value: number): number {
  if (value === 1) return 11;
  if (value >= 10) return 10;
  return value;
}

/** Blackjack hand score with the best (highest non-busting) Ace treatment. */
export function blackjackScore(cards: Card[]): number {
  let total = 0;
  let aces = 0;
  for (const c of cards) {
    total += bjCardBase(c.value);
    if (c.value === 1) aces += 1;
  }
  while (total > 21 && aces > 0) {
    total -= 10;
    aces -= 1;
  }
  return total;
}

/** Preview the two hand totals after swapping the 2nd card between two Blackjack Switch hands. */
export function blackjackSwitchPreviewScores(handA: Card[], handB: Card[]): { a: number; b: number } | null {
  if (handA.length < 2 || handB.length < 2) return null;
  const swappedA = [handA[0], handB[1], ...handA.slice(2)];
  const swappedB = [handB[0], handA[1], ...handB.slice(2)];
  return { a: blackjackScore(swappedA), b: blackjackScore(swappedB) };
}
