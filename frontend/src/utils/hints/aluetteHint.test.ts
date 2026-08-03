import { describe, expect, it } from 'vitest';
import { makeAluetteState } from '../../test/stateFactories';
import { getAluetteHint } from './aluetteHint';

describe('getAluetteHint', () => {
  it('maps a server hint onto the frontend shape', () => {
    const hint = getAluetteHint(makeAluetteState({ hint: { cardIndices: [1], reason: 'play_luette' } }));
    expect(hint).toEqual({ targetAction: 'play', reason: 'hint.play_luette', confidence: 'moderate' });
  });

  it('returns null with no hint at all', () => {
    expect(getAluetteHint(makeAluetteState())).toBeNull();
  });

  it('returns null when the server sent a hint with no reason', () => {
    expect(getAluetteHint(makeAluetteState({ hint: { cardIndices: [0], reason: '' } }))).toBeNull();
  });
});
