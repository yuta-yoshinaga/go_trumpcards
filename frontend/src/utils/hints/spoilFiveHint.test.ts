import { describe, expect, it } from 'vitest';
import { makeSpoilFiveState } from '../../test/stateFactories';
import { getSpoilFiveHint } from './spoilFiveHint';

describe('getSpoilFiveHint', () => {
  it('returns null when there is no hint', () => {
    expect(getSpoilFiveHint(makeSpoilFiveState())).toBeNull();
  });

  it('returns null when the hint has an empty reason', () => {
    expect(getSpoilFiveHint(makeSpoilFiveState({ hint: { cardIndices: [0], reason: '' } }))).toBeNull();
  });

  it('maps a backend hint to a play HintResult with a namespaced reason', () => {
    const result = getSpoilFiveHint(makeSpoilFiveState({ hint: { cardIndices: [1], reason: 'take_trick' } }));
    expect(result).toEqual({ targetAction: 'play', reason: 'hint.take_trick', confidence: 'moderate' });
  });
});
