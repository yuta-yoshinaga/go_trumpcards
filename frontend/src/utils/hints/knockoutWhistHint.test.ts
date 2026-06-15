import { describe, expect, it } from 'vitest';
import { makeKnockoutWhistState } from '../../test/stateFactories';
import { getKnockoutWhistHint } from './knockoutWhistHint';

describe('getKnockoutWhistHint', () => {
  it('returns null when there is no hint', () => {
    expect(getKnockoutWhistHint(makeKnockoutWhistState())).toBeNull();
  });

  it('returns null when the hint has an empty reason', () => {
    expect(getKnockoutWhistHint(makeKnockoutWhistState({ hint: { cardIndices: [0], reason: '' } }))).toBeNull();
  });

  it('maps a backend hint to a play HintResult with a namespaced reason', () => {
    const result = getKnockoutWhistHint(makeKnockoutWhistState({ hint: { cardIndices: [1], reason: 'lead_high' } }));
    expect(result).toEqual({ targetAction: 'play', reason: 'hint.lead_high', confidence: 'moderate' });
  });
});
