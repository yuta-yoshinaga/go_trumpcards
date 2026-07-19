import { describe, expect, it } from 'vitest';
import type { Card } from '../types/card';
import { frenchTarotUnburiableReason } from './frenchtarotEcart';

/** Builds a minimal French Tarot card face for classification tests. */
function card(value: number, color?: string): Card {
  return { design: 'JOKER', value, color } as unknown as Card;
}

describe('frenchTarotUnburiableReason', () => {
  it('flags the Excuse (gold) as un-buriable', () => {
    expect(frenchTarotUnburiableReason(card(0, 'gold'))).toBe('excuse');
  });

  it('flags a suit King (value 14, non-tarot) as un-buriable', () => {
    expect(frenchTarotUnburiableReason(card(14, 'red'))).toBe('king');
    expect(frenchTarotUnburiableReason(card(14, 'black'))).toBe('king');
  });

  it('flags the bouts Petit (trump 1) and the 21 as un-buriable', () => {
    expect(frenchTarotUnburiableReason(card(1, 'purple'))).toBe('bout');
    expect(frenchTarotUnburiableReason(card(21, 'purple'))).toBe('bout');
  });

  it('flags an ordinary trump (purple, non-bout) as un-buriable', () => {
    expect(frenchTarotUnburiableReason(card(7, 'purple'))).toBe('trump');
    // A trump of value 14 is still an ordinary trump, not a King.
    expect(frenchTarotUnburiableReason(card(14, 'purple'))).toBe('trump');
  });

  it('returns null for a freely buriable low suit card', () => {
    expect(frenchTarotUnburiableReason(card(3, 'red'))).toBeNull();
    expect(frenchTarotUnburiableReason(card(13, 'black'))).toBeNull();
  });
});
