import { describe, expect, it } from 'vitest';
import type { Card } from '../types/card';
import { scartoUndiscardableReason } from './scartoDiscard';

/** Builds a minimal Scarto card face for classification tests. */
function card(value: number, color?: string): Card {
  return { design: 'JOKER', value, color } as unknown as Card;
}

describe('scartoUndiscardableReason', () => {
  it('flags the Excuse (gold) as un-buriable', () => {
    expect(scartoUndiscardableReason(card(0, 'gold'))).toBe('excuse');
  });

  it('flags the bouts Petit (trump 1) and the 21 as un-buriable', () => {
    expect(scartoUndiscardableReason(card(1, 'purple'))).toBe('bout');
    expect(scartoUndiscardableReason(card(21, 'purple'))).toBe('bout');
  });

  it('flags an ordinary trump (purple, non-bout) as un-buriable', () => {
    expect(scartoUndiscardableReason(card(7, 'purple'))).toBe('trump');
    // A trump of value 14 is still an ordinary trump, not a court.
    expect(scartoUndiscardableReason(card(14, 'purple'))).toBe('trump');
  });

  it('flags a counting card (court/King, value ≥ 11, non-tarot) as un-buriable', () => {
    expect(scartoUndiscardableReason(card(11, 'red'))).toBe('court');
    expect(scartoUndiscardableReason(card(14, 'black'))).toBe('court');
  });

  it('returns null for a freely buriable low pip', () => {
    expect(scartoUndiscardableReason(card(2, 'red'))).toBeNull();
    expect(scartoUndiscardableReason(card(10, 'black'))).toBeNull();
  });
});
