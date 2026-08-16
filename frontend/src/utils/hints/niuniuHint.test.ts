import { describe, expect, it } from 'vitest';
import type { NiuNiuResponse } from '../../types/card';
import { NiuNiuPhase } from '../../types/phases';
import { getNiuNiuHint } from './niuniuHint';

function base(overrides: Partial<NiuNiuResponse> = {}) {
  return {
    seats: [{ name: 'You', isCpu: false }],
    bankerIdx: 1,
    chips: 300,
    maxMultiplier: 3,
    bankerRankKey: '',
    phase: NiuNiuPhase.BET,
    message: '',
    ...overrides,
  } as unknown as NiuNiuResponse;
}

describe('getNiuNiuHint', () => {
  it('stays quiet outside the betting phase', () => {
    expect(getNiuNiuHint(base({ phase: NiuNiuPhase.END }))).toBeNull();
  });

  // **上限は chips ではなく chips / maxMultiplier。**親の Niu Niu は 3 倍払い。
  it('names the stake cap while the stack covers a bet', () => {
    expect(getNiuNiuHint(base({ chips: 300, maxMultiplier: 3 }))).toEqual({
      targetAction: 'bet',
      reason: 'frontendHint.niuniuStakeCap',
      confidence: 'moderate',
    });
  });

  // 倍率をかけると 1 も賭けられない持ち点。
  it('warns when the stack cannot cover even the smallest stake', () => {
    expect(getNiuNiuHint(base({ chips: 2, maxMultiplier: 3 }))).toEqual({
      targetAction: 'bet',
      reason: 'frontendHint.niuniuStackTooShort',
      confidence: 'strong',
    });
  });

  // ちょうど 1 賭けられる境界。
  it('treats the boundary as playable', () => {
    expect(getNiuNiuHint(base({ chips: 3, maxMultiplier: 3 }))?.reason).toBe('frontendHint.niuniuStakeCap');
  });

  // **倍率が届いていなければ割らない。**0 除算で Infinity を勧めない。
  it('stays quiet without a multiplier', () => {
    expect(getNiuNiuHint(base({ maxMultiplier: 0 }))).toBeNull();
  });

  it('warns on an empty stack', () => {
    expect(getNiuNiuHint(base({ chips: 0 }))?.reason).toBe('frontendHint.niuniuStackTooShort');
  });
});
