import { describe, expect, it } from 'vitest';
import type { Card } from '../types/card';
import { tonkMeldIndices } from './tonkMeldIndices';

const c = (design: string, value: number): Card => ({ design, value }) as unknown as Card;

describe('tonkMeldIndices', () => {
  it('detects a set of three of a kind', () => {
    const cards = [c('SPADE', 7), c('HEART', 7), c('CLOVER', 7), c('DIAMOND', 2)];
    expect(tonkMeldIndices(cards)).toEqual(new Set([0, 1, 2]));
  });

  it('detects a four-of-a-kind set', () => {
    const cards = [c('SPADE', 7), c('HEART', 7), c('CLOVER', 7), c('DIAMOND', 7)];
    expect(tonkMeldIndices(cards)).toEqual(new Set([0, 1, 2, 3]));
  });

  it('detects a same-suit run of three regardless of input order', () => {
    const cards = [c('SPADE', 6), c('SPADE', 4), c('SPADE', 5), c('HEART', 10)];
    expect(tonkMeldIndices(cards)).toEqual(new Set([0, 1, 2]));
  });

  it('does not meld a pair or a two-card partial run', () => {
    expect(tonkMeldIndices([c('SPADE', 7), c('HEART', 7), c('CLOVER', 2)])).toEqual(new Set());
    expect(tonkMeldIndices([c('SPADE', 4), c('SPADE', 5), c('HEART', 9)])).toEqual(new Set());
  });

  it('does not connect a run across different suits', () => {
    expect(tonkMeldIndices([c('SPADE', 4), c('HEART', 5), c('SPADE', 6)])).toEqual(new Set());
  });

  it('detects both a set and a run in the same hand', () => {
    const cards = [c('SPADE', 7), c('HEART', 7), c('CLOVER', 7), c('DIAMOND', 4), c('DIAMOND', 5), c('DIAMOND', 6)];
    expect(tonkMeldIndices(cards)).toEqual(new Set([0, 1, 2, 3, 4, 5]));
  });

  it('returns an empty set for fewer than three cards', () => {
    expect(tonkMeldIndices([c('SPADE', 7), c('HEART', 7)])).toEqual(new Set());
  });
});
