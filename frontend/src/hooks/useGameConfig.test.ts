import { act, renderHook } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import { useGameConfig } from './useGameConfig';

type TestConfig = {
  count: number;
  limit: number;
  enabled: boolean;
};

const DEFAULT: TestConfig = { count: 1, limit: 100, enabled: false };

describe('useGameConfig', () => {
  it('returns default config on init', () => {
    const { result } = renderHook(() => useGameConfig(DEFAULT));
    expect(result.current.config).toEqual(DEFAULT);
  });

  it('handleConfigChange updates config with valid number', () => {
    const { result } = renderHook(() => useGameConfig(DEFAULT));

    act(() => {
      result.current.handleConfigChange('limit', '200');
    });

    expect(result.current.config.limit).toBe(200);
  });

  it('handleConfigChange ignores NaN values', () => {
    const { result } = renderHook(() => useGameConfig(DEFAULT));

    act(() => {
      result.current.handleConfigChange('limit', 'abc');
    });

    expect(result.current.config.limit).toBe(100);
  });

  it('handleToggle updates boolean config', () => {
    const { result } = renderHook(() => useGameConfig(DEFAULT));

    expect(result.current.config.enabled).toBe(false);

    act(() => {
      result.current.handleToggle('enabled', true);
    });

    expect(result.current.config.enabled).toBe(true);
  });

  it('setConfig replaces config directly', () => {
    const { result } = renderHook(() => useGameConfig(DEFAULT));

    act(() => {
      result.current.setConfig({ count: 5, limit: 50, enabled: true });
    });

    expect(result.current.config).toEqual({ count: 5, limit: 50, enabled: true });
  });
});
