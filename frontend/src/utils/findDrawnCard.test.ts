import { describe, expect, it } from 'vitest';
import { findDrawnCard } from './findDrawnCard';

describe('findDrawnCard', () => {
  it('finds a second copy of a card already in the hand', () => {
    const existing = { design: 'HEART' as const, value: 7 };
    const drawn = { design: 'HEART' as const, value: 7 };
    expect(findDrawnCard([existing], [existing, drawn])).toBe(drawn);
  });

  it('returns undefined when no card count increased', () => {
    expect(findDrawnCard([{ design: 'SPADE', value: 3 }], [{ design: 'SPADE', value: 3 }])).toBeUndefined();
  });
});
