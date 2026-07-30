import { renderHook } from '@testing-library/react';
import { afterEach, describe, expect, it } from 'vitest';
import { useTutorialProgress } from './useTutorialProgress';

describe('useTutorialProgress', () => {
  afterEach(() => {
    localStorage.clear();
  });

  it('returns all games with 0 completed by default', () => {
    const { result } = renderHook(() => useTutorialProgress());
    expect(result.current.totalCount).toBe(238);
    expect(result.current.completedCount).toBe(0);
    expect(result.current.games.every((g) => !g.completed)).toBe(true);
  });

  it('detects completed tutorials from localStorage', () => {
    localStorage.setItem('tutorial_completed_blackjack', 'true');
    localStorage.setItem('tutorial_completed_poker', 'true');
    const { result } = renderHook(() => useTutorialProgress());
    expect(result.current.completedCount).toBe(2);
    expect(result.current.games.find((g) => g.gameName === 'blackjack')?.completed).toBe(true);
    expect(result.current.games.find((g) => g.gameName === 'poker')?.completed).toBe(true);
    expect(result.current.games.find((g) => g.gameName === 'hearts')?.completed).toBe(false);
  });

  it('maps root path to blackjack', () => {
    const { result } = renderHook(() => useTutorialProgress());
    const bj = result.current.games.find((g) => g.gameName === 'blackjack');
    expect(bj).toBeDefined();
    expect(bj?.path).toBe('/');
  });

  it('maps non-root paths by removing leading slash', () => {
    const { result } = renderHook(() => useTutorialProgress());
    const poker = result.current.games.find((g) => g.gameName === 'poker');
    expect(poker).toBeDefined();
    expect(poker?.path).toBe('/poker');
  });

  it('includes all game routes with labelKey', () => {
    const { result } = renderHook(() => useTutorialProgress());
    for (const game of result.current.games) {
      expect(game.labelKey).toBeTruthy();
      expect(game.gameName).toBeTruthy();
    }
  });
});
