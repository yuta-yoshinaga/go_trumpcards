import { act, renderHook } from '@testing-library/react';
import { beforeEach, describe, expect, it } from 'vitest';
import { useProfilePersistence } from './useProfilePersistence';

describe('useProfilePersistence', () => {
  beforeEach(() => {
    localStorage.clear();
  });

  it('saveProfile stores data to localStorage', () => {
    const { result } = renderHook(() => useProfilePersistence('poker'));
    const profile = { gamesPlayed: 3, foldToBetCount: 2 };
    act(() => result.current.saveProfile(profile));
    expect(localStorage.getItem('metaai_poker')).toBe(JSON.stringify(profile));
  });

  it('saveProfile is no-op when profile is null', () => {
    const { result } = renderHook(() => useProfilePersistence('poker'));
    act(() => result.current.saveProfile(null));
    expect(localStorage.getItem('metaai_poker')).toBeNull();
  });

  it('saveProfile is no-op when profile is undefined', () => {
    const { result } = renderHook(() => useProfilePersistence('poker'));
    act(() => result.current.saveProfile(undefined));
    expect(localStorage.getItem('metaai_poker')).toBeNull();
  });

  it('loadProfile returns stored data', () => {
    const profile = { gamesPlayed: 5 };
    localStorage.setItem('metaai_doubt', JSON.stringify(profile));
    const { result } = renderHook(() => useProfilePersistence('doubt'));
    expect(result.current.loadProfile()).toEqual(profile);
  });

  it('loadProfile returns undefined when no data stored', () => {
    const { result } = renderHook(() => useProfilePersistence('doubt'));
    expect(result.current.loadProfile()).toBeUndefined();
  });

  it('clearProfile removes data from localStorage', () => {
    localStorage.setItem('metaai_holdem', JSON.stringify({ gamesPlayed: 1 }));
    const { result } = renderHook(() => useProfilePersistence('holdem'));
    act(() => result.current.clearProfile());
    expect(localStorage.getItem('metaai_holdem')).toBeNull();
  });

  it('uses correct key prefix for different games', () => {
    const { result: r1 } = renderHook(() => useProfilePersistence('poker'));
    const { result: r2 } = renderHook(() => useProfilePersistence('holdem'));
    act(() => {
      r1.current.saveProfile({ game: 'poker' });
      r2.current.saveProfile({ game: 'holdem' });
    });
    expect(localStorage.getItem('metaai_poker')).toContain('poker');
    expect(localStorage.getItem('metaai_holdem')).toContain('holdem');
  });
});
