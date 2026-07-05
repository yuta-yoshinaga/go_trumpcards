import { describe, expect, it } from 'vitest';
import { makeUltiState } from '../../test/stateFactories';
import { getUltiHint } from './ultiHint';

describe('getUltiHint', () => {
  it('returns null when the response carries no hint', () => {
    expect(getUltiHint(makeUltiState())).toBeNull();
    expect(getUltiHint(makeUltiState({ hint: null }))).toBeNull();
  });

  it('returns null when the hint reason is empty', () => {
    const state = makeUltiState({ hint: { cardIndices: [], reason: '' } });
    expect(getUltiHint(state)).toBeNull();
  });

  it('maps a server play hint into a HintResult', () => {
    const state = makeUltiState({ hint: { cardIndices: [2], reason: 'follow_win' } });
    expect(getUltiHint(state)).toEqual({
      targetAction: 'play',
      reason: 'hint.follow_win',
      confidence: 'moderate',
    });
  });

  it('maps a lead-high hint reason verbatim', () => {
    const state = makeUltiState({ hint: { cardIndices: [0], reason: 'lead_high' } });
    expect(getUltiHint(state)?.reason).toBe('hint.lead_high');
  });

  it('maps a discard-weak hint reason verbatim', () => {
    const state = makeUltiState({ hint: { cardIndices: [0, 1], reason: 'discard_weak' } });
    expect(getUltiHint(state)?.reason).toBe('hint.discard_weak');
  });

  it('maps a party bid hint reason verbatim', () => {
    const state = makeUltiState({ hint: { cardIndices: [], reason: 'bid_party' } });
    expect(getUltiHint(state)?.reason).toBe('hint.bid_party');
  });
});
