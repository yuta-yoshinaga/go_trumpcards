import { describe, expect, it } from 'vitest';
import { makeMinchiateState } from '../../test/stateFactories';
import { getMinchiateHint } from './minchiateHint';

describe('getMinchiateHint', () => {
  it('maps a server hint onto the frontend shape', () => {
    const hint = getMinchiateHint(makeMinchiateState({ hint: { cardIndices: [1], reason: 'play_papa' } }));
    expect(hint).toEqual({ targetAction: 'play', reason: 'hint.play_papa', confidence: 'moderate' });
  });

  it('returns null with no hint at all', () => {
    expect(getMinchiateHint(makeMinchiateState())).toBeNull();
  });

  it('returns null when the server sent a hint with no reason', () => {
    expect(getMinchiateHint(makeMinchiateState({ hint: { cardIndices: [0], reason: '' } }))).toBeNull();
  });
});
