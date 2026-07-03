import { describe, expect, it } from 'vitest';
import { makeGutsState } from '../../test/stateFactories';
import { getGutsHint } from './gutsHint';

describe('getGutsHint', () => {
  it('returns null when the response carries no hint', () => {
    expect(getGutsHint(makeGutsState())).toBeNull();
    expect(getGutsHint(makeGutsState({ hint: null }))).toBeNull();
  });

  it('returns null when the hint reason is empty', () => {
    const state = makeGutsState({ hint: { declaration: 1, reason: '' } });
    expect(getGutsHint(state)).toBeNull();
  });

  it('maps an in declaration to the in action with a strong-hand reason', () => {
    const state = makeGutsState({ hint: { declaration: 1, reason: 'strong_hand' } });
    expect(getGutsHint(state)).toEqual({
      targetAction: 'in',
      reason: 'hint.strong_hand',
      confidence: 'moderate',
    });
  });

  it('maps an out declaration to the out action with a weak-hand reason', () => {
    const state = makeGutsState({ hint: { declaration: 0, reason: 'weak_hand' } });
    expect(getGutsHint(state)).toEqual({
      targetAction: 'out',
      reason: 'hint.weak_hand',
      confidence: 'moderate',
    });
  });
});
