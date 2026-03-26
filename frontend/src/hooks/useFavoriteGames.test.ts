import { act, renderHook } from '@testing-library/react';
import { afterEach, beforeEach, describe, expect, it } from 'vitest';
import { FAVORITE_GAMES_KEY, useFavoriteGames } from './useFavoriteGames';

describe('useFavoriteGames', () => {
  beforeEach(() => {
    localStorage.removeItem(FAVORITE_GAMES_KEY);
  });

  afterEach(() => {
    localStorage.removeItem(FAVORITE_GAMES_KEY);
  });

  it('returns empty favorites initially', () => {
    const { result } = renderHook(() => useFavoriteGames());
    expect(result.current.favorites).toEqual([]);
  });

  it('isFavorite returns false for unknown path', () => {
    const { result } = renderHook(() => useFavoriteGames());
    expect(result.current.isFavorite('/poker')).toBe(false);
  });

  it('toggleFavorite adds a game', () => {
    const { result } = renderHook(() => useFavoriteGames());
    act(() => {
      result.current.toggleFavorite('/poker');
    });
    expect(result.current.favorites).toContain('/poker');
    expect(result.current.isFavorite('/poker')).toBe(true);
  });

  it('toggleFavorite removes a favorited game', () => {
    const { result } = renderHook(() => useFavoriteGames());
    act(() => {
      result.current.toggleFavorite('/poker');
    });
    act(() => {
      result.current.toggleFavorite('/poker');
    });
    expect(result.current.favorites).not.toContain('/poker');
    expect(result.current.isFavorite('/poker')).toBe(false);
  });

  it('persists favorites to localStorage', () => {
    const { result } = renderHook(() => useFavoriteGames());
    act(() => {
      result.current.toggleFavorite('/hearts');
    });
    const stored = JSON.parse(localStorage.getItem(FAVORITE_GAMES_KEY) || '[]');
    expect(stored).toContain('/hearts');
  });

  it('loads favorites from localStorage on mount', () => {
    localStorage.setItem(FAVORITE_GAMES_KEY, JSON.stringify(['/poker', '/hearts']));
    const { result } = renderHook(() => useFavoriteGames());
    expect(result.current.favorites).toEqual(['/poker', '/hearts']);
    expect(result.current.isFavorite('/poker')).toBe(true);
    expect(result.current.isFavorite('/hearts')).toBe(true);
  });

  it('handles malformed localStorage gracefully', () => {
    localStorage.setItem(FAVORITE_GAMES_KEY, 'not-json');
    const { result } = renderHook(() => useFavoriteGames());
    expect(result.current.favorites).toEqual([]);
  });

  it('filters out stale paths from localStorage', () => {
    localStorage.setItem(FAVORITE_GAMES_KEY, JSON.stringify(['/poker', '/nonexistent', '/hearts']));
    const { result } = renderHook(() => useFavoriteGames());
    expect(result.current.favorites).toEqual(['/poker', '/hearts']);
    expect(result.current.favorites).not.toContain('/nonexistent');
  });

  it('caps favorites at 10', () => {
    const { result } = renderHook(() => useFavoriteGames());
    const paths = [
      '/',
      '/poker',
      '/hearts',
      '/spades',
      '/klondike',
      '/freecell',
      '/spider',
      '/pyramid',
      '/memory',
      '/baccarat',
      '/holdem',
    ];
    for (const p of paths) {
      act(() => {
        result.current.toggleFavorite(p);
      });
    }
    expect(result.current.favorites.length).toBe(10);
    // 11th game should not be added
    expect(result.current.favorites).not.toContain('/holdem');
  });

  it('supports multiple favorites', () => {
    const { result } = renderHook(() => useFavoriteGames());
    act(() => {
      result.current.toggleFavorite('/poker');
    });
    act(() => {
      result.current.toggleFavorite('/hearts');
    });
    act(() => {
      result.current.toggleFavorite('/spades');
    });
    expect(result.current.favorites).toEqual(['/poker', '/hearts', '/spades']);
  });
});
