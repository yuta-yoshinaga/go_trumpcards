import { describe, expect, it } from 'vitest';
import { missingNapoleonRoundCheckpoints } from './napoleonPhaseCoverage';

describe('missingNapoleonRoundCheckpoints', () => {
  it('accepts a complete round when the optional trick-end checkpoint was not observed', () => {
    const observed = new Set(['bid', 'play', 'round-boundary'] as const);

    expect(missingNapoleonRoundCheckpoints(observed)).toEqual([]);
  });

  it('reports a missing bid checkpoint', () => {
    const observed = new Set(['play', 'round-boundary'] as const);

    expect(missingNapoleonRoundCheckpoints(observed)).toEqual(['bid']);
  });

  it('reports all missing checkpoints in required order', () => {
    const observed = new Set(['declare', 'exchange', 'trick-end'] as const);

    expect(missingNapoleonRoundCheckpoints(observed)).toEqual(['bid', 'play', 'round-boundary']);
  });
});
