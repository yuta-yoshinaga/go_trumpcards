import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { createElement } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { anacondaApi } from '../api/gameApi';
import { makeAnacondaState } from '../test/stateFactories';
import { DEFAULT_ANACONDA_CONFIG, useAnacondaGame } from './useAnacondaGame';

vi.mock('../api/gameApi', () => ({
  anacondaApi: { exec: vi.fn() },
  actionLogApi: { anaconda: vi.fn() },
}));

const mockExec = vi.mocked(anacondaApi.exec);

function createWrapper() {
  const queryClient = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  return ({ children }: { children: ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children);
}

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(makeAnacondaState());
});

describe('useAnacondaGame', () => {
  it('reset dispatches reset with the default config', async () => {
    const { result } = renderHook(() => useAnacondaGame(), { wrapper: createWrapper() });
    act(() => result.current.reset());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, DEFAULT_ANACONDA_CONFIG));
  });

  it('handlePass dispatches pass with the indices', async () => {
    const { result } = renderHook(() => useAnacondaGame(), { wrapper: createWrapper() });
    act(() => result.current.handlePass([0, 1, 2]));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('pass', [0, 1, 2]));
  });

  it('handleKeep dispatches keep with the indices', async () => {
    const { result } = renderHook(() => useAnacondaGame(), { wrapper: createWrapper() });
    act(() => result.current.handleKeep([0, 1, 2, 3, 4]));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('keep', [0, 1, 2, 3, 4]));
  });

  it('handleCall dispatches a call bet', async () => {
    const { result } = renderHook(() => useAnacondaGame(), { wrapper: createWrapper() });
    act(() => result.current.handleCall());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bet', undefined, 'call'));
  });

  it('handleRaise dispatches a raise bet', async () => {
    const { result } = renderHook(() => useAnacondaGame(), { wrapper: createWrapper() });
    act(() => result.current.handleRaise());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bet', undefined, 'raise'));
  });

  it('handleFold dispatches a fold bet', async () => {
    const { result } = renderHook(() => useAnacondaGame(), { wrapper: createWrapper() });
    act(() => result.current.handleFold());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bet', undefined, 'fold'));
  });

  it('handleNextRound dispatches nextround', async () => {
    const { result } = renderHook(() => useAnacondaGame(), { wrapper: createWrapper() });
    act(() => result.current.handleNextRound());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
  });

  it('handleConfigChange updates the config used by reset', async () => {
    const { result } = renderHook(() => useAnacondaGame(), { wrapper: createWrapper() });
    act(() => result.current.handleConfigChange('playerCount', '6'));
    act(() => result.current.reset());
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, {
        ...DEFAULT_ANACONDA_CONFIG,
        playerCount: 6,
      }),
    );
  });
});
