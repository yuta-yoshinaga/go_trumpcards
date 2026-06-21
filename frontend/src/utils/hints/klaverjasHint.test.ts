import { describe, expect, it } from 'vitest';
import { makeKlaverjasState } from '../../test/stateFactories';
import { getKlaverjasHint } from './klaverjasHint';

describe('getKlaverjasHint', () => {
  it('returns null when the response carries no hint', () => {
    expect(getKlaverjasHint(makeKlaverjasState())).toBeNull();
    expect(getKlaverjasHint(makeKlaverjasState({ hint: null }))).toBeNull();
  });

  it('returns null when the hint reason is empty', () => {
    const state = makeKlaverjasState({ hint: { cardIndices: [], reason: '' } });
    expect(getKlaverjasHint(state)).toBeNull();
  });

  it('maps a server play hint into a HintResult', () => {
    const state = makeKlaverjasState({ hint: { cardIndices: [2], reason: 'follow_win' } });
    expect(getKlaverjasHint(state)).toEqual({
      targetAction: 'play',
      reason: 'hint.follow_win',
      confidence: 'moderate',
    });
  });

  it('maps a lead hint reason verbatim', () => {
    const state = makeKlaverjasState({ hint: { cardIndices: [0], reason: 'lead_low' } });
    expect(getKlaverjasHint(state)?.reason).toBe('hint.lead_low');
  });

  it('maps a discard hint reason verbatim', () => {
    const state = makeKlaverjasState({ hint: { cardIndices: [1], reason: 'discard_low' } });
    expect(getKlaverjasHint(state)?.reason).toBe('hint.discard_low');
  });
});
