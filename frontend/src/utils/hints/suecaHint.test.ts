import { describe, expect, it } from 'vitest';
import { makeSuecaState } from '../../test/stateFactories';
import { getSuecaHint } from './suecaHint';

describe('getSuecaHint', () => {
  it('returns null when the response carries no hint', () => {
    expect(getSuecaHint(makeSuecaState())).toBeNull();
    expect(getSuecaHint(makeSuecaState({ hint: null }))).toBeNull();
  });

  it('returns null when the hint reason is empty', () => {
    const state = makeSuecaState({ hint: { cardIndices: [], reason: '' } });
    expect(getSuecaHint(state)).toBeNull();
  });

  it('maps a server play hint into a HintResult', () => {
    const state = makeSuecaState({ hint: { cardIndices: [2], reason: 'follow_win' } });
    expect(getSuecaHint(state)).toEqual({
      targetAction: 'play',
      reason: 'hint.follow_win',
      confidence: 'moderate',
    });
  });

  it('maps a lead hint reason verbatim', () => {
    const state = makeSuecaState({ hint: { cardIndices: [0], reason: 'lead_low' } });
    expect(getSuecaHint(state)?.reason).toBe('hint.lead_low');
  });

  it('maps a discard hint reason verbatim', () => {
    const state = makeSuecaState({ hint: { cardIndices: [1], reason: 'discard_low' } });
    expect(getSuecaHint(state)?.reason).toBe('hint.discard_low');
  });
});
