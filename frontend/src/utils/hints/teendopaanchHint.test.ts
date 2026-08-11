import { describe, expect, it } from 'vitest';
import type { TeenDoPaanchResponse } from '../../types/card';
import { getTeenDoPaanchHint } from './teendopaanchHint';

const base = (hint?: TeenDoPaanchResponse['hint']): TeenDoPaanchResponse =>
  ({ hint }) as unknown as TeenDoPaanchResponse;

describe('getTeenDoPaanchHint', () => {
  it('returns null when the server sent no hint', () => {
    expect(getTeenDoPaanchHint(base())).toBeNull();
  });

  // **宣言の助言は札ではなくスートを指す。**
  it('names the suit when it recommends a trump call', () => {
    expect(getTeenDoPaanchHint(base({ reason: 'teendopaanchSelectTrump', suit: 3 }))).toEqual({
      targetAction: 'trump-3',
      reason: 'hint.teendopaanchSelectTrump',
      confidence: 'moderate',
    });
  });

  it('accepts card index 0', () => {
    expect(getTeenDoPaanchHint(base({ cardIndex: 0, reason: 'teendopaanchDuck', suit: 0 }))).toEqual({
      targetAction: 'card-0',
      reason: 'hint.teendopaanchDuck',
      confidence: 'moderate',
    });
  });

  // **ノルマに届いていないときだけ強く勧める。** 届いたら取らないのが正解。
  it('is strong only while short of the target', () => {
    expect(getTeenDoPaanchHint(base({ cardIndex: 4, reason: 'teendopaanchWinTrick', suit: 0 }))?.confidence).toBe(
      'strong',
    );
    expect(getTeenDoPaanchHint(base({ cardIndex: 4, reason: 'teendopaanchDuck', suit: 0 }))?.confidence).toBe(
      'moderate',
    );
  });
});
