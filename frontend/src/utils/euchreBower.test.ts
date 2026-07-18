import { describe, expect, it } from 'vitest';
import type { Card } from '../types/card';
import { bowerRole, sameColorSuit } from './euchreBower';

// suit number → design string, mirroring the domain's CardDesign constants.
const DESIGN: Record<number, Card['design']> = {
  1: 'SPADE',
  2: 'CLOVER',
  3: 'HEART',
  4: 'DIAMOND',
};

const jack = (suit: number): Card => ({ design: DESIGN[suit], value: 11 });

describe('sameColorSuit', () => {
  it('maps black suits to each other (♠↔♣)', () => {
    expect(sameColorSuit(1)).toBe(2);
    expect(sameColorSuit(2)).toBe(1);
  });

  it('maps red suits to each other (♥↔♦)', () => {
    expect(sameColorSuit(3)).toBe(4);
    expect(sameColorSuit(4)).toBe(3);
  });

  it('returns the input unchanged for non-suit values', () => {
    expect(sameColorSuit(0)).toBe(0);
    expect(sameColorSuit(9)).toBe(9);
  });
});

describe('bowerRole', () => {
  // right bower = J of trump; left bower = J of the same-color suit as trump.
  const cases: Array<{ trump: number; right: number; left: number }> = [
    { trump: 1, right: 1, left: 2 }, // ♠ trump → right ♠J, left ♣J
    { trump: 2, right: 2, left: 1 }, // ♣ trump → right ♣J, left ♠J
    { trump: 3, right: 3, left: 4 }, // ♥ trump → right ♥J, left ♦J
    { trump: 4, right: 4, left: 3 }, // ♦ trump → right ♦J, left ♥J
  ];

  for (const { trump, right, left } of cases) {
    it(`identifies right and left bowers for trump suit ${trump}`, () => {
      expect(bowerRole(jack(right), trump)).toBe('right');
      expect(bowerRole(jack(left), trump)).toBe('left');
      // The two off-color jacks are neither bower.
      for (const s of [1, 2, 3, 4]) {
        if (s !== right && s !== left) {
          expect(bowerRole(jack(s), trump)).toBeNull();
        }
      }
    });
  }

  it('returns null for non-Jack cards even in the trump suit', () => {
    expect(bowerRole({ design: 'SPADE', value: 1 }, 1)).toBeNull();
    expect(bowerRole({ design: 'SPADE', value: 13 }, 1)).toBeNull();
  });

  it('returns null when no trump is set', () => {
    expect(bowerRole(jack(1), 0)).toBeNull();
    expect(bowerRole(jack(3), -1)).toBeNull();
  });

  it('returns null for an unknown design string', () => {
    expect(bowerRole({ design: 'JOKER', value: 11 }, 1)).toBeNull();
  });
});
