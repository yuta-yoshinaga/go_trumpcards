import { describe, expect, it } from 'vitest';
import { michiganNextPlayable } from './michiganPlayable';

describe('michiganNextPlayable', () => {
  it('reports a new sequence when needNewSequence is set', () => {
    const result = michiganNextPlayable({
      seqSuit: 2,
      seqSuitName: 'ハート',
      seqHighValue: 5,
      needNewSequence: true,
    });
    expect(result.isNewSequence).toBe(true);
  });

  it('reports a new sequence when seqSuit is 0', () => {
    const result = michiganNextPlayable({
      seqSuit: 0,
      seqSuitName: '',
      seqHighValue: 0,
      needNewSequence: false,
    });
    expect(result.isNewSequence).toBe(true);
  });

  it('computes the next continuation rank for an active sequence', () => {
    const result = michiganNextPlayable({
      seqSuit: 1,
      seqSuitName: 'ハート',
      seqHighValue: 3,
      needNewSequence: false,
    });
    expect(result).toEqual({ isNewSequence: false, suitName: 'ハート', nextValue: 4 });
  });
});
