import { describe, expect, it } from 'vitest';
import { makeBouillotteState } from '../../test/stateFactories';
import { getBouillotteHint } from './bouillotteHint';

describe('getBouillotteHint', () => {
  it('returns null when the response carries no hint', () => {
    expect(getBouillotteHint(makeBouillotteState())).toBeNull();
    expect(getBouillotteHint(makeBouillotteState({ hint: null }))).toBeNull();
  });

  it('returns null when the hint reason is empty', () => {
    const state = makeBouillotteState({ hint: { action: 'call', reason: '' } });
    expect(getBouillotteHint(state)).toBeNull();
  });

  it('maps a raise action with a strong-hand reason', () => {
    const state = makeBouillotteState({ hint: { action: 'raise', reason: 'strong_hand' } });
    expect(getBouillotteHint(state)).toEqual({
      targetAction: 'raise',
      reason: 'hint.strong_hand',
      confidence: 'moderate',
    });
  });

  it('maps a call action with a medium-hand reason', () => {
    const state = makeBouillotteState({ hint: { action: 'call', reason: 'medium_hand' } });
    expect(getBouillotteHint(state)).toEqual({
      targetAction: 'call',
      reason: 'hint.medium_hand',
      confidence: 'moderate',
    });
  });

  it('maps a fold action with a weak-hand reason', () => {
    const state = makeBouillotteState({ hint: { action: 'fold', reason: 'weak_hand' } });
    expect(getBouillotteHint(state)).toEqual({
      targetAction: 'fold',
      reason: 'hint.weak_hand',
      confidence: 'moderate',
    });
  });
});
