import { describe, expect, it } from 'vitest';
import { makeCalabresellaState } from '../../test/stateFactories';
import { getCalabresellaHint } from './calabresellaHint';

describe('getCalabresellaHint', () => {
  it('returns null when the response carries no hint', () => {
    expect(getCalabresellaHint(makeCalabresellaState())).toBeNull();
    expect(getCalabresellaHint(makeCalabresellaState({ hint: null }))).toBeNull();
  });

  it('returns null when the hint reason is empty', () => {
    const state = makeCalabresellaState({ hint: { cardIndices: [], reason: '' } });
    expect(getCalabresellaHint(state)).toBeNull();
  });

  it('maps a server play hint into a HintResult', () => {
    const state = makeCalabresellaState({ hint: { cardIndices: [2], reason: 'follow_win' } });
    expect(getCalabresellaHint(state)).toEqual({
      targetAction: 'play',
      reason: 'hint.follow_win',
      confidence: 'moderate',
    });
  });

  it('maps a lead hint reason verbatim', () => {
    const state = makeCalabresellaState({ hint: { cardIndices: [0], reason: 'lead_low' } });
    expect(getCalabresellaHint(state)?.reason).toBe('hint.lead_low');
  });

  it('maps a give-partner hint reason verbatim', () => {
    const state = makeCalabresellaState({ hint: { cardIndices: [0], reason: 'give_partner' } });
    expect(getCalabresellaHint(state)?.reason).toBe('hint.give_partner');
  });

  it('maps a chiamo bid hint reason verbatim', () => {
    const state = makeCalabresellaState({ hint: { cardIndices: [], reason: 'bid_chiamo' } });
    expect(getCalabresellaHint(state)?.reason).toBe('hint.bid_chiamo');
  });

  it('maps a discard hint reason verbatim', () => {
    const state = makeCalabresellaState({ hint: { cardIndices: [1], reason: 'discard_low' } });
    expect(getCalabresellaHint(state)?.reason).toBe('hint.discard_low');
  });
});
