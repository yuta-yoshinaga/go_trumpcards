import { describe, expect, it } from 'vitest';
import type { Card } from '../types/card';
import golden from './__fixtures__/cinchBidStrength.golden.json';
import { estimateCinchBidStrength } from './cinchBidStrength';

/**
 * The Cinch point rules live twice: `internal/domain/Cinch.go`
 * (`CinchHandPointsBySuit`, which the CUI calls directly) and this module. These
 * golden vectors are also asserted by `TestCinchBidStrength_GoldenVectors` in
 * `internal/domain/Cinch_test.go`, so changing the rules on one side alone fails
 * that side, and regenerating the vectors to fix it fails the other.
 */
const DESIGNS: Card['design'][] = ['JOKER', 'SPADE', 'CLOVER', 'HEART', 'DIAMOND'];

describe('estimateCinchBidStrength golden vectors (shared with the Go domain)', () => {
  it('has vectors to check', () => {
    expect(golden.cases.length).toBeGreaterThan(0);
  });

  it.each(golden.cases)('$name', (c) => {
    const cards: Card[] = c.cards.map((cd) => ({ design: DESIGNS[cd.suit], value: cd.value }));

    const got = estimateCinchBidStrength(cards);

    expect([0, got.pointsBySuit[1], got.pointsBySuit[2], got.pointsBySuit[3], got.pointsBySuit[4]]).toEqual(
      c.pointsBySuit,
    );
    expect(got.bestSuit).toBe(c.bestSuit);
    expect(got.maxPoints).toBe(c.maxPoints);
    expect(got.minPoints).toBe(c.minPoints);
  });
});
