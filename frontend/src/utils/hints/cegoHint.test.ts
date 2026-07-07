import { describe, expect, it } from 'vitest';
import { makeCegoState } from '../../test/stateFactories';
import { getCegoHint } from './cegoHint';

describe('getCegoHint', () => {
  it('returns null when there is no hint', () => {
    expect(getCegoHint(makeCegoState({ hint: null }))).toBeNull();
  });

  it('returns null when the hint has an empty reason', () => {
    expect(getCegoHint(makeCegoState({ hint: { cardIndices: [], reason: '' } }))).toBeNull();
  });

  it('maps a backend play hint to a HintResult', () => {
    const res = getCegoHint(makeCegoState({ hint: { cardIndices: [2], reason: 'lead_high' } }));
    expect(res).toEqual({ targetAction: 'play', reason: 'hint.lead_high', confidence: 'moderate' });
  });

  it('maps a backend contract hint', () => {
    const res = getCegoHint(makeCegoState({ hint: { contract: 1, cardIndices: [], reason: 'contract_cego' } }));
    expect(res).toEqual({ targetAction: 'play', reason: 'hint.contract_cego', confidence: 'moderate' });
  });
});
