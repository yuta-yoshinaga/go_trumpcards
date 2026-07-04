import { describe, expect, it } from 'vitest';
import { makeMichiganState } from '../../test/stateFactories';
import { getMichiganHint } from './michiganHint';

describe('getMichiganHint', () => {
  it('returns null when the response carries no hint', () => {
    expect(getMichiganHint(makeMichiganState())).toBeNull();
    expect(getMichiganHint(makeMichiganState({ hint: null }))).toBeNull();
  });

  it('returns null when the hint reason is empty', () => {
    const state = makeMichiganState({ hint: { cardIndex: 0, reason: '' } });
    expect(getMichiganHint(state)).toBeNull();
  });

  it('maps a forced-card reason to a play hint', () => {
    const state = makeMichiganState({ hint: { cardIndex: 1, reason: 'forced' } });
    expect(getMichiganHint(state)).toEqual({
      targetAction: 'play',
      reason: 'hint.forced',
      confidence: 'moderate',
    });
  });

  it('maps a claim-boodle reason to a play hint', () => {
    const state = makeMichiganState({ hint: { cardIndex: 3, reason: 'claim_boodle' } });
    expect(getMichiganHint(state)).toEqual({
      targetAction: 'play',
      reason: 'hint.claim_boodle',
      confidence: 'moderate',
    });
  });

  it('maps a lead-low reason to a play hint', () => {
    const state = makeMichiganState({ hint: { cardIndex: 2, reason: 'lead_low' } });
    expect(getMichiganHint(state)).toEqual({
      targetAction: 'play',
      reason: 'hint.lead_low',
      confidence: 'moderate',
    });
  });
});
