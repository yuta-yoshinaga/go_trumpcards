import { describe, expect, it } from 'vitest';
import { makeMusState } from '../../test/stateFactories';
import { getMusHint } from './musHint';

describe('getMusHint', () => {
  it('returns null when the response carries no hint', () => {
    expect(getMusHint(makeMusState())).toBeNull();
    expect(getMusHint(makeMusState({ hint: null }))).toBeNull();
  });

  it('returns null when the hint reason is empty', () => {
    const state = makeMusState({ hint: { mus: false, action: 0, amount: 0, indices: [], reason: '' } });
    expect(getMusHint(state)).toBeNull();
  });

  it('maps a mus-phase hint to the mus target action', () => {
    const state = makeMusState({
      phase: 0,
      hint: { mus: true, action: 0, amount: 0, indices: [], reason: 'mus_exchange' },
    });
    expect(getMusHint(state)).toEqual({
      targetAction: 'mus',
      reason: 'hint.mus_exchange',
      confidence: 'moderate',
    });
  });

  it('maps a discard hint to the discard target action', () => {
    const state = makeMusState({
      phase: 1,
      hint: { mus: false, action: 0, amount: 0, indices: [3], reason: 'discard_low' },
    });
    expect(getMusHint(state)?.targetAction).toBe('discard');
    expect(getMusHint(state)?.reason).toBe('hint.discard_low');
  });

  it('maps a bet hint to the bet target action', () => {
    const state = makeMusState({
      hint: { mus: false, action: 1, amount: 2, indices: [], reason: 'bet_envido' },
    });
    expect(getMusHint(state)).toEqual({
      targetAction: 'bet',
      reason: 'hint.bet_envido',
      confidence: 'moderate',
    });
  });
});
