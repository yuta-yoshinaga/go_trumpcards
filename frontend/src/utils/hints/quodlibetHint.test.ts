import { describe, expect, it } from 'vitest';
import { makeQuodlibetState } from '../../test/stateFactories';
import { getQuodlibetHint } from './quodlibetHint';

describe('getQuodlibetHint', () => {
  it('maps a card reason into a hint key', () => {
    expect(getQuodlibetHint(makeQuodlibetState({ hint: { cardIndices: [2], reason: 'avoid_penalty' } }))).toEqual({
      targetAction: 'play',
      reason: 'hint.avoid_penalty',
      confidence: 'moderate',
    });
  });

  // **種目の助言もヒント。** 選択フェーズには指す札が無いので、
  // cardIndices が空なら null という判定だと一番効く助言が消える。
  it('keeps the contract suggestion, which carries no card', () => {
    expect(getQuodlibetHint(makeQuodlibetState({ hint: null, hintContract: 2 }))).toEqual({
      targetAction: 'select',
      reason: 'hint.pick_contract',
      confidence: 'moderate',
    });
  });

  it('returns null when the server sent no hint', () => {
    expect(getQuodlibetHint(makeQuodlibetState())).toBeNull();
    expect(getQuodlibetHint(makeQuodlibetState({ hint: { cardIndices: [], reason: '' } }))).toBeNull();
  });

  it('returns null for the explicit no-suggestion reason', () => {
    expect(getQuodlibetHint(makeQuodlibetState({ hint: { cardIndices: [], reason: 'none' } }))).toBeNull();
  });
});
