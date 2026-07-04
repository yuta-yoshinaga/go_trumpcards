import { describe, expect, it } from 'vitest';
import { makeWattenState } from '../../test/stateFactories';
import { getWattenHint } from './wattenHint';

describe('getWattenHint', () => {
  it('returns null when the response carries no hint', () => {
    expect(getWattenHint(makeWattenState())).toBeNull();
    expect(getWattenHint(makeWattenState({ hint: null }))).toBeNull();
  });

  it('returns null when the hint reason is empty', () => {
    const state = makeWattenState({ hint: { action: 'play', reason: '' } });
    expect(getWattenHint(state)).toBeNull();
  });

  it('maps a server play hint into a HintResult mirroring the action', () => {
    const state = makeWattenState({ hint: { action: 'play', cardIndex: 2, reason: 'lead_trump' } });
    expect(getWattenHint(state)).toEqual({
      targetAction: 'play',
      reason: 'hint.lead_trump',
      confidence: 'moderate',
    });
  });

  it('maps a declare hint with its action', () => {
    const state = makeWattenState({ hint: { action: 'declare', rank: 13, suit: 3, reason: 'declare_strong' } });
    const hint = getWattenHint(state);
    expect(hint?.targetAction).toBe('declare');
    expect(hint?.reason).toBe('hint.declare_strong');
  });

  it('maps a raise hint with its action', () => {
    const state = makeWattenState({ hint: { action: 'raise', reason: 'raise_strong' } });
    expect(getWattenHint(state)?.targetAction).toBe('raise');
    expect(getWattenHint(state)?.reason).toBe('hint.raise_strong');
  });

  it('maps hold and fold hints', () => {
    expect(getWattenHint(makeWattenState({ hint: { action: 'hold', reason: 'hold_ok' } }))?.targetAction).toBe('hold');
    expect(getWattenHint(makeWattenState({ hint: { action: 'fold', reason: 'fold_weak' } }))?.reason).toBe(
      'hint.fold_weak',
    );
  });

  it('falls back to play when the action is empty', () => {
    const state = makeWattenState({ hint: { action: '', reason: 'lead_plain' } });
    expect(getWattenHint(state)?.targetAction).toBe('play');
  });
});
