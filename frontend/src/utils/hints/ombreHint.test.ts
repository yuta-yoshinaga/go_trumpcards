import { describe, expect, it } from 'vitest';
import { makeOmbreState } from '../../test/stateFactories';
import { getOmbreHint } from './ombreHint';

describe('getOmbreHint', () => {
  it('returns null when the response carries no hint', () => {
    expect(getOmbreHint(makeOmbreState())).toBeNull();
    expect(getOmbreHint(makeOmbreState({ hint: null }))).toBeNull();
  });

  it('returns null when the hint reason is empty', () => {
    const state = makeOmbreState({ hint: { cardIndices: [], reason: '' } });
    expect(getOmbreHint(state)).toBeNull();
  });

  it('maps a server play hint into a HintResult', () => {
    const state = makeOmbreState({ hint: { cardIndices: [2], reason: 'follow_win' } });
    expect(getOmbreHint(state)).toEqual({
      targetAction: 'play',
      reason: 'hint.follow_win',
      confidence: 'moderate',
    });
  });

  it('maps a lead-high hint reason verbatim', () => {
    const state = makeOmbreState({ hint: { cardIndices: [0], reason: 'lead_high' } });
    expect(getOmbreHint(state)?.reason).toBe('hint.lead_high');
  });

  it('maps a give-partner hint reason verbatim', () => {
    const state = makeOmbreState({ hint: { cardIndices: [0], reason: 'give_partner' } });
    expect(getOmbreHint(state)?.reason).toBe('hint.give_partner');
  });

  it('maps an entrar bid hint reason verbatim', () => {
    const state = makeOmbreState({ hint: { cardIndices: [], reason: 'bid_entrar' } });
    expect(getOmbreHint(state)?.reason).toBe('hint.bid_entrar');
  });
});
