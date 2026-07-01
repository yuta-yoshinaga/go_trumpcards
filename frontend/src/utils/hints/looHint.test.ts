import { describe, expect, it } from 'vitest';
import { makeLooState } from '../../test/stateFactories';
import { getLooHint } from './looHint';

describe('getLooHint', () => {
  it('returns null when the response carries no hint', () => {
    expect(getLooHint(makeLooState())).toBeNull();
    expect(getLooHint(makeLooState({ hint: null }))).toBeNull();
  });

  it('returns null when the hint reason is empty', () => {
    const state = makeLooState({ hint: { cardIndices: [], reason: '' } });
    expect(getLooHint(state)).toBeNull();
  });

  it('maps a server play hint into a HintResult', () => {
    const state = makeLooState({ hint: { cardIndices: [2], reason: 'follow_win' } });
    expect(getLooHint(state)).toEqual({
      targetAction: 'play',
      reason: 'hint.follow_win',
      confidence: 'moderate',
    });
  });

  it('maps a lead hint reason verbatim', () => {
    const state = makeLooState({ hint: { cardIndices: [0], reason: 'lead_high' } });
    expect(getLooHint(state)?.reason).toBe('hint.lead_high');
  });

  it('maps a decide-play hint reason verbatim', () => {
    const state = makeLooState({ hint: { cardIndices: [], decision: true, reason: 'decide_play' } });
    expect(getLooHint(state)?.reason).toBe('hint.decide_play');
  });

  it('maps a decide-pass hint reason verbatim', () => {
    const state = makeLooState({ hint: { cardIndices: [], decision: false, reason: 'decide_pass' } });
    expect(getLooHint(state)?.reason).toBe('hint.decide_pass');
  });

  it('maps a discard hint reason verbatim', () => {
    const state = makeLooState({ hint: { cardIndices: [3], reason: 'discard_low' } });
    expect(getLooHint(state)?.reason).toBe('hint.discard_low');
  });
});
