import { describe, expect, it } from 'vitest';
import type { SultanResponse } from '../../types/card';
import { getSultanHint } from './sultanHint';

const state = (hint?: SultanResponse['hint']): SultanResponse => ({ hint }) as SultanResponse;

describe('getSultanHint', () => {
  it('returns null without a hint', () => {
    expect(getSultanHint(state())).toBeNull();
  });

  it('names the destination foundation', () => {
    expect(getSultanHint(state({ fromZone: 'divan', fromIdx: 2, toFoundation: 5 }))).toEqual({
      targetAction: 'foundation-5',
      reason: 'frontendHint.sultanMove',
      confidence: 'moderate',
    });
  });

  // **ファウンデーション 0 は正当。**真偽値で見ると先頭だけ落ちる。
  it('keeps a move onto foundation zero', () => {
    expect(getSultanHint(state({ fromZone: 'divan', fromIdx: 1, toFoundation: 0 }))?.targetAction).toBe('foundation-0');
  });
});
