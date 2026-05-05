import { act, renderHook } from '@testing-library/react';
import { afterEach, describe, expect, it } from 'vitest';
import {
  DEFAULT_REPLAY_SPEED,
  getReplaySpeedMultiplier,
  multiplierForSpeed,
  REPLAY_SPEED_STORAGE_KEY,
  useReplaySpeed,
} from './useReplaySpeed';

describe('useReplaySpeed', () => {
  afterEach(() => {
    localStorage.clear();
  });

  it('defaults to normal when nothing is stored', () => {
    const { result } = renderHook(() => useReplaySpeed());
    expect(result.current[0]).toBe(DEFAULT_REPLAY_SPEED);
  });

  it('reads a previously persisted value on mount', () => {
    localStorage.setItem(REPLAY_SPEED_STORAGE_KEY, 'fast');
    const { result } = renderHook(() => useReplaySpeed());
    expect(result.current[0]).toBe('fast');
  });

  it('falls back to default when stored value is invalid', () => {
    localStorage.setItem(REPLAY_SPEED_STORAGE_KEY, 'turbo');
    const { result } = renderHook(() => useReplaySpeed());
    expect(result.current[0]).toBe(DEFAULT_REPLAY_SPEED);
  });

  it('persists changes to localStorage', () => {
    const { result } = renderHook(() => useReplaySpeed());
    act(() => result.current[1]('instant'));
    expect(result.current[0]).toBe('instant');
    expect(localStorage.getItem(REPLAY_SPEED_STORAGE_KEY)).toBe('instant');
  });
});

describe('multiplierForSpeed', () => {
  it.each([
    ['normal', 1],
    ['fast', 0.3],
    ['instant', 0],
  ] as const)('returns %s -> %s', (speed, expected) => {
    expect(multiplierForSpeed(speed)).toBe(expected);
  });
});

describe('getReplaySpeedMultiplier', () => {
  afterEach(() => {
    localStorage.clear();
  });

  it('returns 1 when nothing is stored', () => {
    expect(getReplaySpeedMultiplier()).toBe(1);
  });

  it('reads the live localStorage value', () => {
    localStorage.setItem(REPLAY_SPEED_STORAGE_KEY, 'fast');
    expect(getReplaySpeedMultiplier()).toBe(0.3);
    localStorage.setItem(REPLAY_SPEED_STORAGE_KEY, 'instant');
    expect(getReplaySpeedMultiplier()).toBe(0);
  });

  it('falls back to 1 when stored value is invalid', () => {
    localStorage.setItem(REPLAY_SPEED_STORAGE_KEY, 'turbo');
    expect(getReplaySpeedMultiplier()).toBe(1);
  });
});
