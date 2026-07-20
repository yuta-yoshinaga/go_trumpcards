import { describe, expect, it } from 'vitest';
import type { Card } from '../types/card';
import { classifyRummy500Meld } from './rummy500MeldValidator';

/** Builds a Card with the given suit design and rank value. */
function card(design: Card['design'], value: number): Card {
  return { design, value };
}

describe('classifyRummy500Meld', () => {
  it('accepts a same-rank triple across distinct suits as a set', () => {
    const result = classifyRummy500Meld([card('SPADE', 7), card('HEART', 7), card('CLOVER', 7)]);
    expect(result).toEqual({ valid: true, kind: 'set' });
  });

  it('accepts a four-card set across all four suits', () => {
    const result = classifyRummy500Meld([
      card('SPADE', 10),
      card('HEART', 10),
      card('CLOVER', 10),
      card('DIAMOND', 10),
    ]);
    expect(result).toEqual({ valid: true, kind: 'set' });
  });

  it('accepts a same-suit consecutive run', () => {
    const result = classifyRummy500Meld([card('DIAMOND', 4), card('DIAMOND', 5), card('DIAMOND', 6)]);
    expect(result).toEqual({ valid: true, kind: 'run' });
  });

  it('accepts a low-ace run (A-2-3)', () => {
    const result = classifyRummy500Meld([card('HEART', 1), card('HEART', 2), card('HEART', 3)]);
    expect(result).toEqual({ valid: true, kind: 'run' });
  });

  it('accepts a high-ace run (Q-K-A)', () => {
    const result = classifyRummy500Meld([card('SPADE', 12), card('SPADE', 13), card('SPADE', 1)]);
    expect(result).toEqual({ valid: true, kind: 'run' });
  });

  it('rejects a two-card selection as too few', () => {
    const result = classifyRummy500Meld([card('SPADE', 7), card('HEART', 7)]);
    expect(result).toEqual({ valid: false, kind: null });
  });

  it('rejects a mixed non-meld selection', () => {
    const result = classifyRummy500Meld([card('SPADE', 3), card('HEART', 8), card('CLOVER', 11)]);
    expect(result).toEqual({ valid: false, kind: null });
  });

  it('rejects a five-card same-rank selection (set max is 4)', () => {
    const result = classifyRummy500Meld([
      card('SPADE', 5),
      card('HEART', 5),
      card('CLOVER', 5),
      card('DIAMOND', 5),
      card('SPADE', 5),
    ]);
    expect(result).toEqual({ valid: false, kind: null });
  });

  it('rejects a set with a duplicate suit', () => {
    const result = classifyRummy500Meld([card('SPADE', 9), card('SPADE', 9), card('HEART', 9)]);
    expect(result).toEqual({ valid: false, kind: null });
  });

  it('rejects a run that spans suits', () => {
    const result = classifyRummy500Meld([card('DIAMOND', 4), card('HEART', 5), card('DIAMOND', 6)]);
    expect(result).toEqual({ valid: false, kind: null });
  });

  it('rejects a non-consecutive same-suit selection (gap)', () => {
    const result = classifyRummy500Meld([card('CLOVER', 4), card('CLOVER', 5), card('CLOVER', 7)]);
    expect(result).toEqual({ valid: false, kind: null });
  });

  it('rejects a run with a duplicate rank', () => {
    const result = classifyRummy500Meld([card('CLOVER', 5), card('CLOVER', 5), card('CLOVER', 6)]);
    expect(result).toEqual({ valid: false, kind: null });
  });

  it('accepts a long five-card run', () => {
    const result = classifyRummy500Meld([
      card('CLOVER', 8),
      card('CLOVER', 9),
      card('CLOVER', 10),
      card('CLOVER', 11),
      card('CLOVER', 12),
    ]);
    expect(result).toEqual({ valid: true, kind: 'run' });
  });
});
