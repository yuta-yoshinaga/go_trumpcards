import { describe, expect, it } from 'vitest';
import { makeMarjapussiState } from '../../test/stateFactories';
import { getMarjapussiHint } from './marjapussiHint';

describe('getMarjapussiHint', () => {
  it('returns null when the response carries no hint', () => {
    expect(getMarjapussiHint(makeMarjapussiState())).toBeNull();
    expect(getMarjapussiHint(makeMarjapussiState({ hint: null }))).toBeNull();
  });

  it('returns null when the hint reason is empty', () => {
    const state = makeMarjapussiState({ hint: { cardIndices: [], reason: '' } });
    expect(getMarjapussiHint(state)).toBeNull();
  });

  it('maps a server play hint into a HintResult', () => {
    const state = makeMarjapussiState({ hint: { cardIndices: [2], reason: 'follow_win' } });
    expect(getMarjapussiHint(state)).toEqual({
      targetAction: 'play',
      reason: 'hint.follow_win',
      confidence: 'moderate',
    });
  });

  it('maps a lead hint reason verbatim', () => {
    const state = makeMarjapussiState({ hint: { cardIndices: [0], reason: 'lead_low' } });
    expect(getMarjapussiHint(state)?.reason).toBe('hint.lead_low');
  });

  it('maps a marriage lead hint reason verbatim', () => {
    const state = makeMarjapussiState({ hint: { cardIndices: [0], reason: 'lead_marriage' } });
    expect(getMarjapussiHint(state)?.reason).toBe('hint.lead_marriage');
  });

  it('maps a follow duck hint reason verbatim', () => {
    const state = makeMarjapussiState({ hint: { cardIndices: [1], reason: 'follow_duck' } });
    expect(getMarjapussiHint(state)?.reason).toBe('hint.follow_duck');
  });

  it('maps a discard low hint reason verbatim', () => {
    const state = makeMarjapussiState({ hint: { cardIndices: [2], reason: 'discard_low' } });
    expect(getMarjapussiHint(state)?.reason).toBe('hint.discard_low');
  });
});
