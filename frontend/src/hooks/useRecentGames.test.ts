import { renderHook } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it } from 'vitest';
import { RECENT_GAMES_KEY, useRecentGames } from './useRecentGames';

describe('useRecentGames', () => {
  beforeEach(() => {
    localStorage.removeItem(RECENT_GAMES_KEY);
  });

  afterEach(() => {
    localStorage.removeItem(RECENT_GAMES_KEY);
  });

  it('returns empty array when pathname is not a game route', () => {
    const { result } = renderHook(() => useRecentGames('/unknown'));
    expect(result.current).toEqual([]);
  });

  it('records a game path matching a known route', () => {
    const { result } = renderHook(() => useRecentGames('/poker'));
    expect(result.current).toContain('/poker');
  });

  it('does not record non-game paths', () => {
    const { result } = renderHook(() => useRecentGames('/unknown'));
    expect(result.current).toEqual([]);
  });

  it('deduplicates: most recent comes first', () => {
    const { result, rerender } = renderHook(({ path }) => useRecentGames(path), {
      initialProps: { path: '/poker' },
    });
    rerender({ path: '/hearts' });
    rerender({ path: '/poker' });
    // poker should be first (most recent)
    expect(result.current[0]).toBe('/poker');
    // hearts should still be present
    expect(result.current).toContain('/hearts');
    // No duplicates
    expect(result.current.filter((p: string) => p === '/poker').length).toBe(1);
  });

  it('limits to 5 recent games', () => {
    const paths = ['/poker', '/hearts', '/spades', '/klondike', '/freecell', '/spider'];
    const { result, rerender } = renderHook(({ path }) => useRecentGames(path), {
      initialProps: { path: paths[0] },
    });
    for (let i = 1; i < paths.length; i++) {
      rerender({ path: paths[i] });
    }
    expect(result.current.length).toBe(5);
    // Oldest (poker) should be dropped
    expect(result.current).not.toContain('/poker');
    // Most recent (spider) should be first
    expect(result.current[0]).toBe('/spider');
  });

  it('persists to localStorage', () => {
    renderHook(() => useRecentGames('/poker'));
    const stored = JSON.parse(localStorage.getItem(RECENT_GAMES_KEY) || '[]');
    expect(stored).toContain('/poker');
  });

  it('loads from localStorage on mount', () => {
    localStorage.setItem(RECENT_GAMES_KEY, JSON.stringify(['/hearts', '/poker']));
    const { result } = renderHook(() => useRecentGames('/klondike'));
    expect(result.current).toContain('/hearts');
    expect(result.current).toContain('/poker');
    expect(result.current).toContain('/klondike');
  });

  it('handles malformed localStorage gracefully', () => {
    localStorage.setItem(RECENT_GAMES_KEY, 'not-json');
    const { result } = renderHook(() => useRecentGames('/poker'));
    expect(result.current).toContain('/poker');
  });
});
