import { describe, expect, it } from 'vitest';
import type { FortyAndEightResponse } from '../../types/card';
import { getFortyAndEightHint } from './fortyandeightHint';

const state = (hint?: FortyAndEightResponse['hint']): FortyAndEightResponse => ({ hint }) as FortyAndEightResponse;

describe('getFortyAndEightHint', () => {
  it('returns null without a hint', () => {
    expect(getFortyAndEightHint(state())).toBeNull();
  });

  it('names the destination column', () => {
    const s = state({ fromZone: 'tableau', fromCol: 1, cardIndex: 0, toZone: 'tableau', toCol: 5 });
    expect(getFortyAndEightHint(s)).toEqual({
      targetAction: 'tableau-5',
      reason: 'frontendHint.fortyandeightMove',
      confidence: 'moderate',
    });
  });

  // **列 0 は正当な列。**真偽値で見ると先頭の山だけ落ちる。
  it('keeps a move onto column zero', () => {
    const s = state({ fromZone: 'waste', fromCol: -1, cardIndex: 0, toZone: 'tableau', toCol: 0 });
    expect(getFortyAndEightHint(s)?.targetAction).toBe('tableau-0');
  });

  // 列を持たないゾーンは -1 で届く。連結すると foundation--1 になる。
  it('names a column-less destination by zone alone', () => {
    const s = state({ fromZone: 'tableau', fromCol: 2, cardIndex: 0, toZone: 'foundation', toCol: -1 });
    expect(getFortyAndEightHint(s)?.targetAction).toBe('foundation');
  });
});
