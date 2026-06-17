import { describe, expect, it } from 'vitest';
import { makeTeenPattiState } from '../../test/stateFactories';
import { getTeenPattiHint } from './teenPattiHint';

describe('getTeenPattiHint', () => {
  it('returns null when the response carries no hint', () => {
    expect(getTeenPattiHint(makeTeenPattiState())).toBeNull();
    expect(getTeenPattiHint(makeTeenPattiState({ hint: null }))).toBeNull();
  });

  it('returns null when the hint reason is empty', () => {
    const state = makeTeenPattiState({ hint: { action: 'bet', reason: '' } });
    expect(getTeenPattiHint(state)).toBeNull();
  });

  it('maps a bet hint into a bet HintResult', () => {
    const state = makeTeenPattiState({ hint: { action: 'bet', reason: 'bet' } });
    expect(getTeenPattiHint(state)).toEqual({
      targetAction: 'bet',
      reason: 'hint.bet',
      confidence: 'moderate',
    });
  });

  it('maps a fold hint to the fold action', () => {
    const state = makeTeenPattiState({ hint: { action: 'fold', reason: 'fold' } });
    expect(getTeenPattiHint(state)).toEqual({
      targetAction: 'fold',
      reason: 'hint.fold',
      confidence: 'moderate',
    });
  });

  it('maps a sideshow hint reason verbatim', () => {
    const state = makeTeenPattiState({ hint: { action: 'sideshow', reason: 'sideshow' } });
    expect(getTeenPattiHint(state)?.reason).toBe('hint.sideshow');
  });

  it('falls back to the bet action when the hint omits an action', () => {
    const state = makeTeenPattiState({ hint: { action: '', reason: 'see' } });
    expect(getTeenPattiHint(state)?.targetAction).toBe('bet');
  });
});
