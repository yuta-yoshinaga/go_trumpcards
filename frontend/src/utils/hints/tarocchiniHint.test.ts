import { describe, expect, it } from 'vitest';
import { makeTarocchiniState } from '../../test/stateFactories';
import { getTarocchiniHint } from './tarocchiniHint';

describe('getTarocchiniHint', () => {
  it('maps a server hint onto the frontend shape', () => {
    const hint = getTarocchiniHint(makeTarocchiniState({ hint: { cardIndices: [1], reason: 'play_papa' } }));
    expect(hint).toEqual({ targetAction: 'play', reason: 'hint.play_papa', confidence: 'moderate' });
  });

  it('returns null with no hint at all', () => {
    expect(getTarocchiniHint(makeTarocchiniState())).toBeNull();
  });

  it('returns null when the server sent a hint with no reason', () => {
    expect(getTarocchiniHint(makeTarocchiniState({ hint: { cardIndices: [0], reason: '' } }))).toBeNull();
  });
});
