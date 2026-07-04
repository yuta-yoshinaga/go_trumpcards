import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { createElement } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { primeroApi } from '../api/gameApi';
import { makePrimeroState } from '../test/stateFactories';
import { DEFAULT_PRIMERO_CONFIG, usePrimeroGame } from './usePrimeroGame';

vi.mock('../api/gameApi', () => ({
  primeroApi: { exec: vi.fn() },
  actionLogApi: { primero: vi.fn() },
}));

const mockExec = vi.mocked(primeroApi.exec);

function createWrapper() {
  const queryClient = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  return ({ children }: { children: ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children);
}

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(makePrimeroState());
});

describe('usePrimeroGame', () => {
  it('reset dispatches reset with the default config', async () => {
    const { result } = renderHook(() => usePrimeroGame(), { wrapper: createWrapper() });
    act(() => result.current.reset());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, DEFAULT_PRIMERO_CONFIG));
  });

  it('handleCall dispatches bet with call', async () => {
    const { result } = renderHook(() => usePrimeroGame(), { wrapper: createWrapper() });
    act(() => result.current.handleCall());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bet', 'call'));
  });

  it('handleRaise dispatches bet with raise', async () => {
    const { result } = renderHook(() => usePrimeroGame(), { wrapper: createWrapper() });
    act(() => result.current.handleRaise());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bet', 'raise'));
  });

  it('handleFold dispatches bet with fold', async () => {
    const { result } = renderHook(() => usePrimeroGame(), { wrapper: createWrapper() });
    act(() => result.current.handleFold());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bet', 'fold'));
  });

  it('handleNextRound dispatches nextround', async () => {
    const { result } = renderHook(() => usePrimeroGame(), { wrapper: createWrapper() });
    act(() => result.current.handleNextRound());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
  });

  it('handleConfigChange updates the config used by reset', async () => {
    const { result } = renderHook(() => usePrimeroGame(), { wrapper: createWrapper() });
    act(() => result.current.handleConfigChange('playerCount', '3'));
    act(() => result.current.reset());
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, { ...DEFAULT_PRIMERO_CONFIG, playerCount: 3 }),
    );
  });
});
