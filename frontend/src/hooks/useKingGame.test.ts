import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { createElement } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { kingApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { makeKingState } from '../test/stateFactories';
import { DEFAULT_KING_CONFIG, useKingGame } from './useKingGame';

vi.mock('../api/gameApi', () => ({
  kingApi: { exec: vi.fn() },
  actionLogApi: { king: vi.fn() },
}));

const mockExec = vi.mocked(kingApi.exec);

function createWrapper() {
  const queryClient = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  return ({ children }: { children: ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children);
}

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(makeKingState());
});

describe('useKingGame', () => {
  it('reset dispatches reset with the default config', async () => {
    const { result } = renderHook(() => useKingGame(), { wrapper: createWrapper() });
    act(() => result.current.reset());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', { config: DEFAULT_KING_CONFIG }));
  });

  it('selectContract dispatches contract with a default trumpSuit of -1', async () => {
    const { result } = renderHook(() => useKingGame(), { wrapper: createWrapper() });
    act(() => result.current.selectContract(2));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('contract', { contract: 2, trumpSuit: -1 }));
  });

  it('selectContract forwards an explicit trump suit for the King (Trump) contract', async () => {
    const { result } = renderHook(() => useKingGame(), { wrapper: createWrapper() });
    act(() => result.current.selectContract(6, 3));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('contract', { contract: 6, trumpSuit: 3 }));
  });

  it('handlePlay does nothing without exactly one selected card', async () => {
    const { result } = renderHook(() => useKingGame(), { wrapper: createWrapper() });
    act(() => result.current.handlePlay());
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('handlePlay dispatches play for the single selected card', async () => {
    const { result } = renderHook(() => useKingGame(), { wrapper: createWrapper() });
    act(() => result.current.toggleCard(2));
    act(() => result.current.handlePlay());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { handIndex: 2 }));
  });

  it('handleNextDeal and hint dispatch their commands', async () => {
    const { result } = renderHook(() => useKingGame(), { wrapper: createWrapper() });
    act(() => result.current.handleNextDeal());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
    act(() => result.current.hint());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('hint'));
  });

  it('handleConfigChange updates the config used by reset', async () => {
    const { result } = renderHook(() => useKingGame(), { wrapper: createWrapper() });
    act(() => result.current.handleConfigChange('cpuDifficulty', '2'));
    act(() => result.current.reset());
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', { config: { ...DEFAULT_KING_CONFIG, cpuDifficulty: 2 } }),
    );
  });
});
