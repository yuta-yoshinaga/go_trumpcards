import { describe, expect, it } from 'vitest';
import { makePreferenceState } from '../../test/stateFactories';
import { getPreferenceHint } from './preferenceHint';

describe('getPreferenceHint', () => {
  it('returns null when the response carries no hint', () => {
    expect(getPreferenceHint(makePreferenceState())).toBeNull();
    expect(getPreferenceHint(makePreferenceState({ hint: null }))).toBeNull();
  });

  it('returns null when the hint reason is empty', () => {
    const state = makePreferenceState({ hint: { cardIndices: [], reason: '' } });
    expect(getPreferenceHint(state)).toBeNull();
  });

  it('maps a server play hint into a HintResult', () => {
    const state = makePreferenceState({ hint: { cardIndices: [2], reason: 'follow_win' } });
    expect(getPreferenceHint(state)).toEqual({
      targetAction: 'play',
      reason: 'hint.follow_win',
      confidence: 'moderate',
    });
  });

  it('maps a lead hint reason verbatim', () => {
    const state = makePreferenceState({ hint: { cardIndices: [0], reason: 'lead_low' } });
    expect(getPreferenceHint(state)?.reason).toBe('hint.lead_low');
  });

  it('maps a discard hint reason verbatim', () => {
    const state = makePreferenceState({ hint: { cardIndices: [1], reason: 'discard_low' } });
    expect(getPreferenceHint(state)?.reason).toBe('hint.discard_low');
  });
});
