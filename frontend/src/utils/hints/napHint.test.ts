import { describe, expect, it } from 'vitest';
import { makeNapState } from '../../test/stateFactories';
import { getNapHint } from './napHint';

describe('getNapHint', () => {
  it('returns null when the response carries no hint', () => {
    expect(getNapHint(makeNapState())).toBeNull();
    expect(getNapHint(makeNapState({ hint: null }))).toBeNull();
  });

  it('returns null when the hint reason is empty', () => {
    const state = makeNapState({ hint: { cardIndices: [], reason: '' } });
    expect(getNapHint(state)).toBeNull();
  });

  it('maps a server play hint into a HintResult', () => {
    const state = makeNapState({ hint: { cardIndices: [2], reason: 'follow_win' } });
    expect(getNapHint(state)).toEqual({
      targetAction: 'play',
      reason: 'hint.follow_win',
      confidence: 'moderate',
    });
  });

  it('maps a lead hint reason verbatim', () => {
    const state = makeNapState({ hint: { cardIndices: [0], reason: 'lead_high' } });
    expect(getNapHint(state)?.reason).toBe('hint.lead_high');
  });

  it('maps a discard hint reason verbatim', () => {
    const state = makeNapState({ hint: { cardIndices: [1], reason: 'discard_low' } });
    expect(getNapHint(state)?.reason).toBe('hint.discard_low');
  });
});
