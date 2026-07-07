import { describe, expect, it } from 'vitest';
import { makeScartoState } from '../../test/stateFactories';
import { getScartoHint } from './scartoHint';

describe('getScartoHint', () => {
  it('returns null when there is no hint', () => {
    expect(getScartoHint(makeScartoState({ hint: null }))).toBeNull();
    expect(getScartoHint(makeScartoState({ hint: undefined }))).toBeNull();
  });

  it('returns null when the hint has an empty reason', () => {
    expect(getScartoHint(makeScartoState({ hint: { cardIndices: [1], reason: '' } }))).toBeNull();
  });

  it('maps a play hint reason into a namespaced HintResult', () => {
    const result = getScartoHint(makeScartoState({ hint: { cardIndices: [2], reason: 'lead_low' } }));
    expect(result).toEqual({ targetAction: 'play', reason: 'hint.lead_low', confidence: 'moderate' });
  });

  it('maps a scarto hint reason', () => {
    const result = getScartoHint(makeScartoState({ hint: { cardIndices: [0, 1, 2], reason: 'scarto_weak' } }));
    expect(result).toEqual({ targetAction: 'play', reason: 'hint.scarto_weak', confidence: 'moderate' });
  });
});
