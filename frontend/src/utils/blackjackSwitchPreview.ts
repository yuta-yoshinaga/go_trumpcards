import type { Card } from '../types/card';

/** Numeric value contributed by one card in Blackjack scoring. Aces report 11; soft-total adjustment happens in {@link blackjackScore}. */
function bjCardBase(value: number): number {
  if (value === 1) return 11;
  if (value >= 10) return 10;
  return value;
}

/**
 * Blackjack hand score with the best (highest non-busting) Ace treatment. Null entries
 * (face-down cards in the server's `(Card | null)[]` payload) are skipped.
 */
export function blackjackScore(cards: ReadonlyArray<Card | null>): number {
  let total = 0;
  let aces = 0;
  for (const c of cards) {
    if (c === null) continue;
    total += bjCardBase(c.value);
    if (c.value === 1) aces += 1;
  }
  while (total > 21 && aces > 0) {
    total -= 10;
    aces -= 1;
  }
  return total;
}

/**
 * Preview the two hand totals after swapping the 2nd card between two Blackjack Switch hands.
 * Returns `null` when either hand has fewer than 2 cards or the swap slot is face-down.
 */
export function blackjackSwitchPreviewScores(
  handA: ReadonlyArray<Card | null>,
  handB: ReadonlyArray<Card | null>,
): { a: number; b: number } | null {
  if (handA.length < 2 || handB.length < 2) return null;
  const a1 = handA[1];
  const b1 = handB[1];
  if (a1 === null || b1 === null) return null;
  const swappedA: Array<Card | null> = [handA[0], b1, ...handA.slice(2)];
  const swappedB: Array<Card | null> = [handB[0], a1, ...handB.slice(2)];
  return { a: blackjackScore(swappedA), b: blackjackScore(swappedB) };
}
