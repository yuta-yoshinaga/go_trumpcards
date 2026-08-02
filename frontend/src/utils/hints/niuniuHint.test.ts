import { describe, expect, it } from 'vitest';
import type { NiuNiuResponse } from '../../types/card';
import { NiuNiuPhase } from '../../types/phases';
import { getNiuNiuHint } from './niuniuHint';

function base(overrides: Partial<NiuNiuResponse> = {}) {
  return {
    seats: [],
    bankerIdx: 0,
    chips: 2000,
    maxMultiplier: 3,
    lastResult: '',
    phase: NiuNiuPhase.BET,
    ...overrides,
  } as NiuNiuResponse;
}

describe('getNiuNiuHint', () => {
  it('stays quiet once the round is settled', () => {
    expect(getNiuNiuHint(base({ phase: NiuNiuPhase.END }))).toBeNull();
  });

  it('stays quiet until a multiplier has been set', () => {
    expect(getNiuNiuHint(base({ maxMultiplier: 0 }))).toBeNull();
  });

  it('offers the largest stake when the stack covers it', () => {
    // 500 × 3 = 1,500 ≦ 2,000。
    const hint = getNiuNiuHint(base({ chips: 2000 }));
    expect(hint?.targetAction).toBe('bet-500');
    expect(hint?.reason).toBe('frontendHint.niuniuBetMax');
  });

  it('measures affordability against the multiplier, not the stake', () => {
    // チップ 600。額面だけ見れば 500 が買えるが、露出は 1,500 になる。
    const hint = getNiuNiuHint(base({ chips: 600 }));
    expect(hint?.targetAction).toBe('bet-100');
    expect(hint?.reason).toBe('frontendHint.niuniuBetCapped');
  });

  it('treats an exactly-covered stake as affordable', () => {
    // 100 × 3 = 300 ちょうど。「未満」で見ると一段下がってしまう。
    expect(getNiuNiuHint(base({ chips: 300 }))?.targetAction).toBe('bet-100');
  });

  it('says the stack cannot back even the smallest stake', () => {
    // チップ 29。倍率 3 では 10 の賭けにも 30 が要る。
    const hint = getNiuNiuHint(base({ chips: 29 }));
    expect(hint?.reason).toBe('frontendHint.niuniuCannotCover');
  });

  it('follows the multiplier when it changes', () => {
    // 同じ 1,500 チップでも倍率 3 なら 500、倍率 10 なら 100 が上限。
    expect(getNiuNiuHint(base({ chips: 1500, maxMultiplier: 3 }))?.targetAction).toBe('bet-500');
    expect(getNiuNiuHint(base({ chips: 1500, maxMultiplier: 10 }))?.targetAction).toBe('bet-100');
  });
});
