import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { createElement } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { tysiacApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { makeTysiacState } from '../test/stateFactories';
import { DEFAULT_TYSIAC_CONFIG, useTysiacGame } from './useTysiacGame';

vi.mock('../api/gameApi', () => ({
  tysiacApi: { exec: vi.fn() },
  actionLogApi: { tysiac: vi.fn() },
}));

const mockExec = vi.mocked(tysiacApi.exec);

function createWrapper() {
  const queryClient = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  return ({ children }: { children: ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children);
}

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(makeTysiacState());
});

describe('useTysiacGame', () => {
  it('reset dispatches reset with the default config', async () => {
    const { result } = renderHook(() => useTysiacGame(), { wrapper: createWrapper() });
    act(() => result.current.reset());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', { config: DEFAULT_TYSIAC_CONFIG }));
  });

  it('handleBid dispatches bid with the raise flag', async () => {
    const { result } = renderHook(() => useTysiacGame(), { wrapper: createWrapper() });
    act(() => result.current.handleBid(true));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', { raise: true }));
    act(() => result.current.handleBid(false));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', { raise: false }));
  });

  it('handleDiscard does nothing without exactly one selected card', async () => {
    const { result } = renderHook(() => useTysiacGame(), { wrapper: createWrapper() });
    act(() => result.current.handleDiscard());
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('handleDiscard dispatches discard for the single selected card', async () => {
    const { result } = renderHook(() => useTysiacGame(), { wrapper: createWrapper() });
    act(() => result.current.toggleCard(1));
    act(() => result.current.handleDiscard());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('discard', { cardIndex: 1 }));
  });

  it('handlePlay does nothing without exactly one selected card', async () => {
    const { result } = renderHook(() => useTysiacGame(), { wrapper: createWrapper() });
    act(() => result.current.handlePlay());
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('handlePlay dispatches play for the single selected card', async () => {
    const { result } = renderHook(() => useTysiacGame(), { wrapper: createWrapper() });
    act(() => result.current.toggleCard(2));
    act(() => result.current.handlePlay());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { cardIndex: 2 }));
  });

  it('handleNextTrick and handleNextRound dispatch their commands', async () => {
    const { result } = renderHook(() => useTysiacGame(), { wrapper: createWrapper() });
    act(() => result.current.handleNextTrick());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
    act(() => result.current.handleNextRound());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
  });

  it('handleConfigChange updates the config used by reset', async () => {
    const { result } = renderHook(() => useTysiacGame(), { wrapper: createWrapper() });
    act(() => result.current.handleConfigChange('targetPoints', '1500'));
    act(() => result.current.reset());
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', { config: { ...DEFAULT_TYSIAC_CONFIG, targetPoints: 1500 } }),
    );
  });
});
