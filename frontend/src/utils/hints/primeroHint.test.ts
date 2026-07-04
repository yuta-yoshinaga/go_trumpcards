import { describe, expect, it } from 'vitest';
import { makePrimeroState } from '../../test/stateFactories';
import { getPrimeroHint } from './primeroHint';

describe('getPrimeroHint', () => {
  it('returns null when the response carries no hint', () => {
    expect(getPrimeroHint(makePrimeroState())).toBeNull();
    expect(getPrimeroHint(makePrimeroState({ hint: null }))).toBeNull();
  });

  it('returns null when the hint reason is empty', () => {
    const state = makePrimeroState({ hint: { action: 'call', reason: '' } });
    expect(getPrimeroHint(state)).toBeNull();
  });

  it('maps a raise action with a strong-hand reason', () => {
    const state = makePrimeroState({ hint: { action: 'raise', reason: 'strong_hand' } });
    expect(getPrimeroHint(state)).toEqual({
      targetAction: 'raise',
      reason: 'hint.strong_hand',
      confidence: 'moderate',
    });
  });

  it('maps a call action with a medium-hand reason', () => {
    const state = makePrimeroState({ hint: { action: 'call', reason: 'medium_hand' } });
    expect(getPrimeroHint(state)).toEqual({
      targetAction: 'call',
      reason: 'hint.medium_hand',
      confidence: 'moderate',
    });
  });

  it('maps a fold action with a weak-hand reason', () => {
    const state = makePrimeroState({ hint: { action: 'fold', reason: 'weak_hand' } });
    expect(getPrimeroHint(state)).toEqual({
      targetAction: 'fold',
      reason: 'hint.weak_hand',
      confidence: 'moderate',
    });
  });
});
