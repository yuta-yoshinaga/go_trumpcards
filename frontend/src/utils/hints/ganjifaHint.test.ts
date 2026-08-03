import { describe, expect, it } from 'vitest';
import { makeGanjifaState } from '../../test/stateFactories';
import { getGanjifaHint } from './ganjifaHint';

describe('getGanjifaHint', () => {
  it('maps a server hint onto the frontend shape', () => {
    const hint = getGanjifaHint(makeGanjifaState({ hint: { cardIndices: [1], reason: 'follow_win' } }));
    expect(hint).toEqual({ targetAction: 'play', reason: 'hint.follow_win', confidence: 'moderate' });
  });

  it('returns null with no hint at all', () => {
    expect(getGanjifaHint(makeGanjifaState())).toBeNull();
  });

  it('returns null when the server sent a hint with no reason', () => {
    expect(getGanjifaHint(makeGanjifaState({ hint: { cardIndices: [0], reason: '' } }))).toBeNull();
  });
});
