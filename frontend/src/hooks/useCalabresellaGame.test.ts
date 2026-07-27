import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { createElement } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { calabresellaApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { makeCalabresellaState } from '../test/stateFactories';
import { DEFAULT_CALABRESELLA_CONFIG, useCalabresellaGame } from './useCalabresellaGame';

vi.mock('../api/gameApi', () => ({
  calabresellaApi: { exec: vi.fn() },
  actionLogApi: { calabresella: vi.fn() },
}));

const mockExec = vi.mocked(calabresellaApi.exec);

function createWrapper() {
  const queryClient = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  return ({ children }: { children: ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children);
}

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(makeCalabresellaState());
});

describe('useCalabresellaGame', () => {
  it('reset dispatches reset with the default config', async () => {
    const { result } = renderHook(() => useCalabresellaGame(), { wrapper: createWrapper() });
    act(() => result.current.reset());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', { config: DEFAULT_CALABRESELLA_CONFIG }));
  });

  it('handleBid dispatches bid with the bid value', async () => {
    const { result } = renderHook(() => useCalabresellaGame(), { wrapper: createWrapper() });
    act(() => result.current.handleBid(1));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', { bid: 1 }));
    act(() => result.current.handleBid(0));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', { bid: 0 }));
  });

  it('handleDiscard does nothing without exactly one selected card', async () => {
    const { result } = renderHook(() => useCalabresellaGame(), { wrapper: createWrapper() });
    act(() => result.current.handleDiscard());
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('handleDiscard dispatches discard for the single selected card', async () => {
    const { result } = renderHook(() => useCalabresellaGame(), { wrapper: createWrapper() });
    act(() => result.current.toggleCard(1));
    act(() => result.current.handleDiscard());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('discard', { cardIndex: 1 }));
  });

  it('handlePlay does nothing without exactly one selected card', async () => {
    const { result } = renderHook(() => useCalabresellaGame(), { wrapper: createWrapper() });
    act(() => result.current.handlePlay());
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('handlePlay dispatches play for the single selected card', async () => {
    const { result } = renderHook(() => useCalabresellaGame(), { wrapper: createWrapper() });
    act(() => result.current.toggleCard(2));
    act(() => result.current.handlePlay());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { cardIndex: 2 }));
  });

  it('handleNextTrick and handleNextRound dispatch their commands', async () => {
    const { result } = renderHook(() => useCalabresellaGame(), { wrapper: createWrapper() });
    act(() => result.current.handleNextTrick());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
    act(() => result.current.handleNextRound());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
  });

  it('handleConfigChange updates the config used by reset', async () => {
    const { result } = renderHook(() => useCalabresellaGame(), { wrapper: createWrapper() });
    act(() => result.current.handleConfigChange('targetPoints', '31'));
    act(() => result.current.reset());
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', { config: { ...DEFAULT_CALABRESELLA_CONFIG, targetPoints: 31 } }),
    );
  });
});
