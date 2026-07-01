import { describe, expect, it } from 'vitest';
import { makeKingState } from '../../test/stateFactories';
import { getKingHint } from './kingHint';

describe('getKingHint', () => {
  it('returns null when the response carries no hint', () => {
    expect(getKingHint(makeKingState())).toBeNull();
    expect(getKingHint(makeKingState({ hint: null }))).toBeNull();
  });

  it('returns null when the hint reason is empty', () => {
    const state = makeKingState({ hint: { cardIndices: [], reason: '' } });
    expect(getKingHint(state)).toBeNull();
  });

  it('maps an avoid-low server hint into a HintResult', () => {
    const state = makeKingState({ hint: { cardIndices: [2], reason: 'avoid_low' } });
    expect(getKingHint(state)).toEqual({
      targetAction: 'play',
      reason: 'hint.avoid_low',
      confidence: 'moderate',
    });
  });

  it('maps a win-high hint reason verbatim', () => {
    const state = makeKingState({ hint: { cardIndices: [0], reason: 'win_high' } });
    expect(getKingHint(state)?.reason).toBe('hint.win_high');
  });
});
