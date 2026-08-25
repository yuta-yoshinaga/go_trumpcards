import { describe, expect, it } from 'vitest';
import { makeDehlaPakadState } from '../../test/stateFactories';
import { getDehlaPakadHint } from './dehlaPakadHint';

describe('getDehlaPakadHint', () => {
  it('maps a card reason into a hint key', () => {
    expect(getDehlaPakadHint(makeDehlaPakadState({ hint: { cardIndices: [2], reason: 'take_the_ten' } }))).toEqual({
      targetAction: 'play',
      reason: 'hint.take_the_ten',
      confidence: 'moderate',
    });
  });

  // **切り札の助言もヒント。** 宣言フェーズには指す札が無いので、cardIndices が
  // 空なら null という判定だと、ハンド全体を決める助言が消える。
  it('keeps the trump suggestion, which carries no card', () => {
    expect(getDehlaPakadHint(makeDehlaPakadState({ hint: null, hintTrumpSuit: 3 }))).toEqual({
      targetAction: 'select',
      reason: 'hint.call_longest',
      confidence: 'moderate',
    });
  });

  it('returns null when the server sent no hint', () => {
    expect(getDehlaPakadHint(makeDehlaPakadState())).toBeNull();
    expect(getDehlaPakadHint(makeDehlaPakadState({ hint: { cardIndices: [], reason: '' } }))).toBeNull();
  });

  it('returns null for the explicit no-suggestion reason', () => {
    expect(getDehlaPakadHint(makeDehlaPakadState({ hint: { cardIndices: [], reason: 'none' } }))).toBeNull();
  });
});
