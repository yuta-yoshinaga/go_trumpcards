import { describe, expect, it } from 'vitest';
import { makeBeziqueState } from '../../test/stateFactories';
import { getBeziqueHint } from './beziqueHint';

describe('getBeziqueHint', () => {
  it('returns null when the response carries no hint', () => {
    expect(getBeziqueHint(makeBeziqueState())).toBeNull();
    expect(getBeziqueHint(makeBeziqueState({ hint: null }))).toBeNull();
  });

  it('returns null when the hint reason is empty', () => {
    const state = makeBeziqueState({ hint: { reason: '' } });
    expect(getBeziqueHint(state)).toBeNull();
  });

  it('maps a server play hint into a play HintResult', () => {
    const state = makeBeziqueState({ hint: { cardIndex: 2, reason: 'follow_cut' } });
    expect(getBeziqueHint(state)).toEqual({
      targetAction: 'play',
      reason: 'hint.follow_cut',
      confidence: 'moderate',
    });
  });

  it('maps a meld-declare hint to the meld action', () => {
    const state = makeBeziqueState({ hint: { meldIndex: 0, reason: 'meld_declare' } });
    expect(getBeziqueHint(state)).toEqual({
      targetAction: 'meld',
      reason: 'hint.meld_declare',
      confidence: 'moderate',
    });
  });

  it('maps a meld-skip hint (meldIndex -1) to the meld action', () => {
    const state = makeBeziqueState({ hint: { meldIndex: -1, reason: 'meld_skip' } });
    expect(getBeziqueHint(state)).toEqual({
      targetAction: 'meld',
      reason: 'hint.meld_skip',
      confidence: 'moderate',
    });
  });

  it('maps a lead_trump hint reason verbatim', () => {
    const state = makeBeziqueState({ hint: { cardIndex: 0, reason: 'lead_trump' } });
    expect(getBeziqueHint(state)?.reason).toBe('hint.lead_trump');
  });

  it('maps a follow_dump hint reason verbatim', () => {
    const state = makeBeziqueState({ hint: { cardIndex: 3, reason: 'follow_dump' } });
    expect(getBeziqueHint(state)?.reason).toBe('hint.follow_dump');
  });
});
