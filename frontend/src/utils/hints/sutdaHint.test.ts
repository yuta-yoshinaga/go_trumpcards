import { describe, expect, it } from 'vitest';
import { makeSutdaState } from '../../test/stateFactories';
import { getSutdaHint } from './sutdaHint';

describe('getSutdaHint', () => {
  // **指す札が無い。** ソッタの判断はすべてベットなので、助言は行動に乗る。
  it.each([
    ['raise', 'strong_hand'],
    ['fold', 'weak_hand'],
    ['call', 'stay_in'],
  ])('carries the %s suggestion', (hintAction, hintReason) => {
    expect(getSutdaHint(makeSutdaState({ hintAction, hintReason }))).toEqual({
      targetAction: hintAction,
      reason: `hint.${hintReason}`,
      confidence: 'moderate',
    });
  });

  it('returns null when the server sent no suggestion', () => {
    expect(getSutdaHint(makeSutdaState())).toBeNull();
    expect(getSutdaHint(makeSutdaState({ hintAction: 'call', hintReason: '' }))).toBeNull();
    expect(getSutdaHint(makeSutdaState({ hintAction: '', hintReason: 'stay_in' }))).toBeNull();
  });

  it('returns null for the explicit no-suggestion reason', () => {
    expect(getSutdaHint(makeSutdaState({ hintAction: 'call', hintReason: 'none' }))).toBeNull();
  });
});
