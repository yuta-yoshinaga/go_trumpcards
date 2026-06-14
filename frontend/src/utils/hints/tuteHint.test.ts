import { describe, expect, it } from 'vitest';
import { makeTuteState } from '../../test/stateFactories';
import { getTuteHint } from './tuteHint';

describe('getTuteHint', () => {
  it('returns null when the response carries no hint', () => {
    expect(getTuteHint(makeTuteState())).toBeNull();
    expect(getTuteHint(makeTuteState({ hint: null }))).toBeNull();
  });

  it('returns null when the hint reason is empty', () => {
    const state = makeTuteState({ hint: { cardIndices: [], marriage: 0, reason: '' } });
    expect(getTuteHint(state)).toBeNull();
  });

  it('maps a server play hint into a HintResult', () => {
    const state = makeTuteState({
      hint: { cardIndices: [2], marriage: 0, reason: 'follow_win' },
    });
    expect(getTuteHint(state)).toEqual({
      targetAction: 'play',
      reason: 'hint.follow_win',
      confidence: 'moderate',
    });
  });

  it('maps a lead hint reason verbatim', () => {
    const state = makeTuteState({
      hint: { cardIndices: [0], marriage: 0, reason: 'lead_low' },
    });
    expect(getTuteHint(state)?.reason).toBe('hint.lead_low');
  });

  it('maps a marriage hint reason verbatim', () => {
    const state = makeTuteState({
      hint: { cardIndices: [], marriage: 3, reason: 'declare_marriage' },
    });
    expect(getTuteHint(state)?.reason).toBe('hint.declare_marriage');
  });
});
