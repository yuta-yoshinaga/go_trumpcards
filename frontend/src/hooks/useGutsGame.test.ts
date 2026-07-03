import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { createElement } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { gutsApi } from '../api/gameApi';
import { makeGutsState } from '../test/stateFactories';
import { DEFAULT_GUTS_CONFIG, useGutsGame } from './useGutsGame';

vi.mock('../api/gameApi', () => ({
  gutsApi: { exec: vi.fn() },
  actionLogApi: { guts: vi.fn() },
}));

const mockExec = vi.mocked(gutsApi.exec);

function createWrapper() {
  const queryClient = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  return ({ children }: { children: ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children);
}

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(makeGutsState());
});

describe('useGutsGame', () => {
  it('reset dispatches reset with the default config', async () => {
    const { result } = renderHook(() => useGutsGame(), { wrapper: createWrapper() });
    act(() => result.current.reset());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, DEFAULT_GUTS_CONFIG));
  });

  it('handleIn dispatches declare with 1', async () => {
    const { result } = renderHook(() => useGutsGame(), { wrapper: createWrapper() });
    act(() => result.current.handleIn());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('declare', 1));
  });

  it('handleOut dispatches declare with 0', async () => {
    const { result } = renderHook(() => useGutsGame(), { wrapper: createWrapper() });
    act(() => result.current.handleOut());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('declare', 0));
  });

  it('handleNextRound dispatches nextround', async () => {
    const { result } = renderHook(() => useGutsGame(), { wrapper: createWrapper() });
    act(() => result.current.handleNextRound());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
  });

  it('handleConfigChange updates the config used by reset', async () => {
    const { result } = renderHook(() => useGutsGame(), { wrapper: createWrapper() });
    act(() => result.current.handleConfigChange('playerCount', '6'));
    act(() => result.current.reset());
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, { ...DEFAULT_GUTS_CONFIG, playerCount: 6 }),
    );
  });
});
