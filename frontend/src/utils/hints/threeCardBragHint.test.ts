import { describe, expect, it } from 'vitest';
import { makeThreeCardBragState } from '../../test/stateFactories';
import { getThreeCardBragHint } from './threeCardBragHint';

describe('getThreeCardBragHint', () => {
  it('returns null when the response carries no hint', () => {
    expect(getThreeCardBragHint(makeThreeCardBragState())).toBeNull();
    expect(getThreeCardBragHint(makeThreeCardBragState({ hint: null }))).toBeNull();
  });

  it('returns null when the hint reason is empty', () => {
    const state = makeThreeCardBragState({ hint: { action: 'bet', reason: '' } });
    expect(getThreeCardBragHint(state)).toBeNull();
  });

  it('maps a bet hint into a bet HintResult', () => {
    const state = makeThreeCardBragState({ hint: { action: 'bet', reason: 'bet' } });
    expect(getThreeCardBragHint(state)).toEqual({
      targetAction: 'bet',
      reason: 'hint.bet',
      confidence: 'moderate',
    });
  });

  it('maps a fold hint to the fold action', () => {
    const state = makeThreeCardBragState({ hint: { action: 'fold', reason: 'fold' } });
    expect(getThreeCardBragHint(state)).toEqual({
      targetAction: 'fold',
      reason: 'hint.fold',
      confidence: 'moderate',
    });
  });

  it('maps a raise hint reason verbatim', () => {
    const state = makeThreeCardBragState({ hint: { action: 'raise', reason: 'raise' } });
    expect(getThreeCardBragHint(state)?.reason).toBe('hint.raise');
  });

  it('falls back to the bet action when the hint omits an action', () => {
    const state = makeThreeCardBragState({ hint: { action: '', reason: 'see' } });
    expect(getThreeCardBragHint(state)?.targetAction).toBe('bet');
  });
});
