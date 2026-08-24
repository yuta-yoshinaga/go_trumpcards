import { describe, expect, it } from 'vitest';
import { makeGermanSoloState } from '../../test/stateFactories';
import { getGermanSoloHint } from './germansoloHint';

describe('getGermanSoloHint', () => {
  it('returns null when the response carries no hint', () => {
    expect(getGermanSoloHint(makeGermanSoloState())).toBeNull();
    expect(getGermanSoloHint(makeGermanSoloState({ hint: null }))).toBeNull();
  });

  it('returns null when the hint reason is empty', () => {
    const state = makeGermanSoloState({ hint: { cardIndices: [], reason: '' } });
    expect(getGermanSoloHint(state)).toBeNull();
  });

  it('maps a server play hint into a HintResult', () => {
    const state = makeGermanSoloState({ hint: { cardIndices: [2], reason: 'follow_win' } });
    expect(getGermanSoloHint(state)).toEqual({
      targetAction: 'play',
      reason: 'hint.follow_win',
      confidence: 'moderate',
    });
  });

  it('maps a lead-high hint reason verbatim', () => {
    const state = makeGermanSoloState({ hint: { cardIndices: [0], reason: 'lead_high' } });
    expect(getGermanSoloHint(state)?.reason).toBe('hint.lead_high');
  });

  it('maps a give-partner hint reason verbatim', () => {
    const state = makeGermanSoloState({ hint: { cardIndices: [0], reason: 'give_partner' } });
    expect(getGermanSoloHint(state)?.reason).toBe('hint.give_partner');
  });

  it('maps an entrar bid hint reason verbatim', () => {
    const state = makeGermanSoloState({ hint: { cardIndices: [], reason: 'bid_entrar' } });
    expect(getGermanSoloHint(state)?.reason).toBe('hint.bid_entrar');
  });
});
