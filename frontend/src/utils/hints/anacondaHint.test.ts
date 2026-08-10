import { describe, expect, it } from 'vitest';
import { makeAnacondaState } from '../../test/stateFactories';
import { getAnacondaHint } from './anacondaHint';

describe('getAnacondaHint', () => {
  it('returns null when the response carries no hint', () => {
    expect(getAnacondaHint(makeAnacondaState())).toBeNull();
    expect(getAnacondaHint(makeAnacondaState({ hint: null }))).toBeNull();
  });

  it('returns null when the hint reason is empty', () => {
    const state = makeAnacondaState({ hint: { action: 'pass', reason: '' } });
    expect(getAnacondaHint(state)).toBeNull();
  });

  it('maps a pass suggestion to the pass action with a moderate confidence', () => {
    const state = makeAnacondaState({ hint: { action: 'pass', cardIndices: [4, 5, 6], reason: 'pass_weakest' } });
    // **どの札かまで返す。**バックエンドは計算済みなのに捨てていた (#4851)。
    expect(getAnacondaHint(state)).toEqual({
      targetAction: 'pass',
      reason: 'hint.pass_weakest',
      confidence: 'moderate',
      targetIndices: [4, 5, 6],
    });
  });

  it('maps a keep suggestion to the keep action', () => {
    const state = makeAnacondaState({ hint: { action: 'keep', cardIndices: [0, 1, 2, 3, 4], reason: 'keep_best' } });
    expect(getAnacondaHint(state)).toEqual({
      targetAction: 'keep',
      reason: 'hint.keep_best',
      confidence: 'moderate',
      targetIndices: [0, 1, 2, 3, 4],
    });
  });

  it('maps a raise suggestion with a strong-hand reason to strong confidence', () => {
    const state = makeAnacondaState({ hint: { action: 'raise', reason: 'strong_hand' } });
    expect(getAnacondaHint(state)).toEqual({
      targetAction: 'raise',
      reason: 'hint.strong_hand',
      confidence: 'strong',
    });
  });

  it('maps a fold suggestion with a weak-hand reason to moderate confidence', () => {
    const state = makeAnacondaState({ hint: { action: 'fold', reason: 'weak_hand' } });
    expect(getAnacondaHint(state)).toEqual({
      targetAction: 'fold',
      reason: 'hint.weak_hand',
      confidence: 'moderate',
    });
  });

  // call/raise/fold では backend が cardIndices を入れないので、キー自体を持たない。
  it('omits targetIndices when the suggestion is a betting action', () => {
    for (const action of ['raise', 'call', 'fold']) {
      const r = getAnacondaHint(makeAnacondaState({ hint: { action, reason: 'medium_hand' } }));
      expect(r).not.toHaveProperty('targetIndices');
    }
    // 空配列でも付けない (リングを 0 個描くだけの無駄な差分を避ける)。
    const empty = getAnacondaHint(
      makeAnacondaState({ hint: { action: 'pass', cardIndices: [], reason: 'weak_hand' } }),
    );
    expect(empty).not.toHaveProperty('targetIndices');
  });
});
