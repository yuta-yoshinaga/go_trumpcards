/**
 * @vitest-environment jsdom
 */
import { renderHook } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import type { UserMood } from '../utils/recommendationScoring';
import { useGameRecommendations } from './useGameRecommendations';

const ALL_SKIP: UserMood = {
  mood: [null, null],
  skill: [null, null],
  social: [null, null],
  theme: [null, null],
};

describe('useGameRecommendations', () => {
  it('returns top3, stretch, and also against the real gameRoutes', () => {
    const { result } = renderHook(() => useGameRecommendations(ALL_SKIP));
    expect(result.current.top3).toHaveLength(3);
    expect(result.current.also.length).toBeGreaterThan(0);
    expect(result.current.also.length).toBeLessThanOrEqual(7);
  });

  it('memoizes — re-rendering with the same mood returns the same object', () => {
    const { result, rerender } = renderHook(({ m }: { m: UserMood }) => useGameRecommendations(m), {
      initialProps: { m: ALL_SKIP },
    });
    const first = result.current;
    rerender({ m: ALL_SKIP });
    expect(result.current).toBe(first);
  });

  it('recomputes when the mood changes', () => {
    const { result, rerender } = renderHook(({ m }: { m: UserMood }) => useGameRecommendations(m), {
      initialProps: { m: ALL_SKIP },
    });
    const first = result.current;
    rerender({ m: { ...ALL_SKIP, mood: [0, 0] } });
    expect(result.current).not.toBe(first);
  });
});
