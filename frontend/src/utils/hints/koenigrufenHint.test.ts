import { describe, expect, it } from 'vitest';
import { makeKoenigrufenState } from '../../test/stateFactories';
import { getKoenigrufenHint } from './koenigrufenHint';

describe('getKoenigrufenHint', () => {
  it('returns null when there is no hint', () => {
    expect(getKoenigrufenHint(makeKoenigrufenState({ hint: null }))).toBeNull();
  });

  it('returns null when the hint has an empty reason', () => {
    expect(getKoenigrufenHint(makeKoenigrufenState({ hint: { cardIndices: [], reason: '' } }))).toBeNull();
  });

  it('maps a backend play hint to a HintResult', () => {
    const res = getKoenigrufenHint(makeKoenigrufenState({ hint: { cardIndices: [2], reason: 'lead_high' } }));
    expect(res).toEqual({ targetAction: 'play', reason: 'hint.lead_high', confidence: 'moderate' });
  });

  it('maps a backend call-king hint', () => {
    const res = getKoenigrufenHint(
      makeKoenigrufenState({ hint: { callSuit: 3, cardIndices: [], reason: 'call_king' } }),
    );
    expect(res).toEqual({ targetAction: 'play', reason: 'hint.call_king', confidence: 'moderate' });
  });
});
