import type { Card } from '../types/card';

/** Returns a Baloot card's points for the declared mode and trump suit. */
export function balootCardPoints(card: Card | null | undefined, mode: number, trumpSuit: number): number {
  if (!card) return 0;
  if (mode === 2 && card.design === ({ 1: 'SPADE', 2: 'CLOVER', 3: 'HEART', 4: 'DIAMOND' } as const)[trumpSuit]) {
    return ({ 9: 14, 10: 10, 11: 20, 12: 3, 13: 4, 1: 11 } as Record<number, number>)[card.value] ?? 0;
  }
  return ({ 1: 11, 10: 10, 11: 2, 12: 3, 13: 4 } as Record<number, number>)[card.value] ?? 0;
}
