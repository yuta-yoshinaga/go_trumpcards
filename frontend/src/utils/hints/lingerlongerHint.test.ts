import { describe, expect, it } from 'vitest';
import type { LingerLongerResponse } from '../../types/card';
import { getLingerLongerHint } from './lingerlongerHint';

const state = (hint?: LingerLongerResponse['hint']): LingerLongerResponse => ({ hint }) as LingerLongerResponse;

describe('getLingerLongerHint', () => {
  it('returns null without a hint', () => {
    expect(getLingerLongerHint(state())).toBeNull();
  });

  // **取れば補充できる場面だけは確実。** それ以外は読みの問題。
  it('is certain when taking the trick refills your hand', () => {
    expect(getLingerLongerHint(state({ cardIndex: 1, reason: 'lingerlongerWinTrick' }))).toEqual({
      targetAction: 'card-1',
      reason: 'hint.lingerlongerWinTrick',
      confidence: 'strong',
    });
  });

  // **山札が空なら取っても補充は無い。** 助言の確度が下がる。
  it.each(['lingerlongerNoStock', 'lingerlongerDuck'])('is only moderate for %s', (reason) => {
    expect(getLingerLongerHint(state({ cardIndex: 2, reason }))).toEqual({
      targetAction: 'card-2',
      reason: `hint.${reason}`,
      confidence: 'moderate',
    });
  });

  // cardIndex 0 は「札を指していない」ではない。
  it('treats card index zero as a card', () => {
    expect(getLingerLongerHint(state({ cardIndex: 0, reason: 'lingerlongerDuck' }))?.targetAction).toBe('card-0');
  });

  // 出せる札が無いときはサーバが札を指さない。
  it('returns null when the hint names no card', () => {
    expect(getLingerLongerHint(state({ reason: 'lingerlongerDuck' }))).toBeNull();
  });
});
