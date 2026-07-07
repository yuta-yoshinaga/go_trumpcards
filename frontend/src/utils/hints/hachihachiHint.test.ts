import { describe, expect, it } from 'vitest';
import { makeHachiHachiState } from '../../test/stateFactories';
import { getHachiHachiHint } from './hachihachiHint';

describe('getHachiHachiHint', () => {
  it('returns null when the response carries no hint', () => {
    expect(getHachiHachiHint(makeHachiHachiState())).toBeNull();
    expect(getHachiHachiHint(makeHachiHachiState({ hint: null }))).toBeNull();
  });

  it('returns null when the hint reason is empty', () => {
    const state = makeHachiHachiState({ hint: { cardIndex: 0, fieldIndex: 0, reason: '' } });
    expect(getHachiHachiHint(state)).toBeNull();
  });

  it('maps a server capture hint into a HintResult targeting play', () => {
    const state = makeHachiHachiState({ hint: { cardIndex: 1, fieldIndex: 0, reason: 'capture' } });
    expect(getHachiHachiHint(state)).toEqual({
      targetAction: 'play',
      reason: 'hint.capture',
      confidence: 'moderate',
    });
  });

  it('maps a discard_low hint', () => {
    const state = makeHachiHachiState({ hint: { cardIndex: 2, fieldIndex: -1, reason: 'discard_low' } });
    expect(getHachiHachiHint(state)).toEqual({
      targetAction: 'play',
      reason: 'hint.discard_low',
      confidence: 'moderate',
    });
  });
});
