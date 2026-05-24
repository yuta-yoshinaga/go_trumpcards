import { describe, expect, it } from 'vitest';
import type { Card } from '../types/card';
import { computeGapsGhostHint } from './gapsGhostHint';

const c = (design: Card['design'], value: number): Card => ({ design, value });

describe('computeGapsGhostHint', () => {
  it('returns anySuit/2 for column 0', () => {
    expect(computeGapsGhostHint([null, c('SPADE', 5)], 0)).toEqual({ kind: 'anySuit', value: 2 });
  });

  it('returns needed (left.suit, left.value + 1) for a normal gap', () => {
    expect(computeGapsGhostHint([c('SPADE', 5), null], 1)).toEqual({
      kind: 'needed',
      design: 'SPADE',
      value: 6,
    });
  });

  it('returns blocked when the left neighbor is a King', () => {
    expect(computeGapsGhostHint([c('HEART', 13), null], 1)).toEqual({ kind: 'blocked' });
  });

  it('returns null when the left neighbor is itself empty (chained gap)', () => {
    expect(computeGapsGhostHint([null, null, c('CLOVER', 4)], 1)).toBeNull();
  });

  it('returns Q-then-K = needed K (value 13)', () => {
    expect(computeGapsGhostHint([c('DIAMOND', 12), null], 1)).toEqual({
      kind: 'needed',
      design: 'DIAMOND',
      value: 13,
    });
  });
});
