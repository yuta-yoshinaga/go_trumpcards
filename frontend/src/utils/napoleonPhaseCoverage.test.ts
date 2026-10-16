import { describe, expect, it } from 'vitest';
import { missingNapoleonRoundCheckpoints } from './napoleonPhaseCoverage';

describe('missingNapoleonRoundCheckpoints', () => {
  it('accepts a complete round when the optional trick-end checkpoint was not observed', () => {
    const required = new Set(['bid', 'play', 'round-boundary'] as const);
    const observed = new Set(['bid', 'play', 'round-boundary'] as const);

    expect(missingNapoleonRoundCheckpoints(required, observed)).toEqual([]);
  });

  it('reports a missing bid checkpoint', () => {
    const required = new Set(['bid', 'play', 'round-boundary'] as const);
    const observed = new Set(['play', 'round-boundary'] as const);

    expect(missingNapoleonRoundCheckpoints(required, observed)).toEqual(['bid']);
  });
});
