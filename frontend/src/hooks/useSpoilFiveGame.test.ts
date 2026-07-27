import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { createElement } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { spoilFiveApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { makeSpoilFiveState } from '../test/stateFactories';
import { DEFAULT_SPOIL_FIVE_CONFIG, useSpoilFiveGame } from './useSpoilFiveGame';

vi.mock('../api/gameApi', () => ({
  spoilFiveApi: { exec: vi.fn() },
  actionLogApi: { spoilfive: vi.fn() },
}));

const mockExec = vi.mocked(spoilFiveApi.exec);

function createWrapper() {
  const queryClient = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  return ({ children }: { children: ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children);
}

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(makeSpoilFiveState());
});

describe('useSpoilFiveGame', () => {
  it('reset dispatches reset with the default config', async () => {
    const { result } = renderHook(() => useSpoilFiveGame(), { wrapper: createWrapper() });
    act(() => result.current.reset());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', { config: DEFAULT_SPOIL_FIVE_CONFIG }));
  });

  it('handlePlay does nothing without exactly one selected card', async () => {
    const { result } = renderHook(() => useSpoilFiveGame(), { wrapper: createWrapper() });
    act(() => result.current.handlePlay());
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('handlePlay dispatches play for the single selected card', async () => {
    const { result } = renderHook(() => useSpoilFiveGame(), { wrapper: createWrapper() });
    act(() => result.current.toggleCard(2));
    act(() => result.current.handlePlay());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { cardIndex: 2 }));
  });

  it('handleNextTrick and handleNextRound dispatch their commands', async () => {
    const { result } = renderHook(() => useSpoilFiveGame(), { wrapper: createWrapper() });
    act(() => result.current.handleNextTrick());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
    act(() => result.current.handleNextRound());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
  });

  it('handleConfigChange updates the config used by reset', async () => {
    const { result } = renderHook(() => useSpoilFiveGame(), { wrapper: createWrapper() });
    act(() => result.current.handleConfigChange('cpuDifficulty', '2'));
    act(() => result.current.reset());
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', {
        config: { ...DEFAULT_SPOIL_FIVE_CONFIG, cpuDifficulty: 2 },
      }),
    );
  });
});
