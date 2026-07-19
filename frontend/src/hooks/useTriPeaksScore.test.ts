import { renderHook } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { useTriPeaksScore } from './useTriPeaksScore';

describe('useTriPeaksScore', () => {
  it('starts at zero', () => {
    const { result } = renderHook(() => useTriPeaksScore(0, 24, 0));
    expect(result.current).toEqual({ score: 0, chain: 0 });
  });

  it('stays zero while inputs are undefined', () => {
    const { result } = renderHook(() => useTriPeaksScore(undefined, undefined, 0));
    expect(result.current).toEqual({ score: 0, chain: 0 });
  });

  it('accumulates a chain across consecutive removals', () => {
    const { result, rerender } = renderHook(
      ({ m, s, p }: { m: number; s: number; p: number }) => useTriPeaksScore(m, s, p),
      { initialProps: { m: 0, s: 24, p: 0 } },
    );
    rerender({ m: 1, s: 24, p: 0 }); // 1st removal → +100
    expect(result.current).toEqual({ score: 100, chain: 1 });
    rerender({ m: 2, s: 24, p: 0 }); // 2nd removal → +200
    expect(result.current).toEqual({ score: 300, chain: 2 });
  });

  it('breaks the chain on a draw but keeps the score', () => {
    const { result, rerender } = renderHook(
      ({ m, s, p }: { m: number; s: number; p: number }) => useTriPeaksScore(m, s, p),
      { initialProps: { m: 0, s: 24, p: 0 } },
    );
    rerender({ m: 1, s: 24, p: 0 }); // +100, chain 1
    rerender({ m: 1, s: 23, p: 0 }); // draw → chain 0
    expect(result.current).toEqual({ score: 100, chain: 0 });
    rerender({ m: 2, s: 23, p: 0 }); // removal restarts chain at 1 → +100
    expect(result.current).toEqual({ score: 200, chain: 1 });
  });
});
