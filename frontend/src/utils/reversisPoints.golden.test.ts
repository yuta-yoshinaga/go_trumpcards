import { describe, expect, it } from 'vitest';
import golden from './__fixtures__/reversisPoints.golden.json';
import { reversisCardPoints } from './reversisPoints';

/**
 * The penalty table lives twice: `ReversisCardPenalty` in
 * `internal/domain/Reversis.go` (which decides the score) and this module (which
 * labels the hand). `TestReversisCardPenalty_GoldenVectors` asserts the same
 * vectors from the Go side, so changing one alone fails that side.
 */
describe('reversisCardPoints golden vectors (shared with the Go domain)', () => {
  it('has vectors to check', () => {
    expect(golden.cases.length).toBeGreaterThan(0);
  });

  it.each(golden.cases)('$name', (c) => {
    expect(reversisCardPoints({ design: 'SPADE', value: c.value })).toBe(c.points);
  });

  it('scores nothing for a missing card', () => {
    expect(reversisCardPoints(null)).toBe(0);
  });
});
