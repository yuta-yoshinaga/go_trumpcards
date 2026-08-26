import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { createElement } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { gleekApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { makeGleekState } from '../test/stateFactories';
import { DEFAULT_GLEEK_CONFIG, useGleekGame } from './useGleekGame';

vi.mock('../api/gameApi', () => ({
  gleekApi: { exec: vi.fn() },
  actionLogApi: { gleek: vi.fn() },
}));

const mockExec = vi.mocked(gleekApi.exec);

function createWrapper() {
  const queryClient = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  return ({ children }: { children: ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children);
}

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(makeGleekState());
});

describe('useGleekGame', () => {
  it('reset dispatches reset with the default config', async () => {
    const { result } = renderHook(() => useGleekGame(), { wrapper: createWrapper() });
    act(() => result.current.reset());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', { config: DEFAULT_GLEEK_CONFIG }));
  });

  it('handleBid dispatches a raise', async () => {
    const { result } = renderHook(() => useGleekGame(), { wrapper: createWrapper() });
    act(() => result.current.handleBid(14));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', { bid: 14 }));
  });

  it('handleBid dispatches dropping out as bid 0', async () => {
    const { result } = renderHook(() => useGleekGame(), { wrapper: createWrapper() });
    act(() => result.current.handleBid(0));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', { bid: 0 }));
  });

  // **捨て札フェーズを抜ける唯一の操作。** これが無いと落札の直後に play が
  // 「フェーズが違う」で弾かれ続ける。
  it('handleDiscard dispatches every selected card', async () => {
    const { result } = renderHook(() => useGleekGame(), { wrapper: createWrapper() });
    act(() => {
      result.current.toggleCard(1);
      result.current.toggleCard(3);
    });
    act(() => result.current.handleDiscard());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('discard', { discardIndices: [1, 3] }));
  });

  it('handleDiscard does nothing with nothing selected', async () => {
    const { result } = renderHook(() => useGleekGame(), { wrapper: createWrapper() });
    act(() => result.current.handleDiscard());
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('handlePlay does nothing without exactly one selected card', async () => {
    const { result } = renderHook(() => useGleekGame(), { wrapper: createWrapper() });
    act(() => result.current.handlePlay());
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('handlePlay dispatches play for the single selected card', async () => {
    const { result } = renderHook(() => useGleekGame(), { wrapper: createWrapper() });
    act(() => result.current.toggleCard(2));
    act(() => result.current.handlePlay());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { cardIndex: 2 }));
  });

  it('handleNextTrick and handleNextRound dispatch their commands', async () => {
    const { result } = renderHook(() => useGleekGame(), { wrapper: createWrapper() });
    act(() => result.current.handleNextTrick());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
    act(() => result.current.handleNextRound());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
  });

  it('handleConfigChange updates the config used by reset', async () => {
    const { result } = renderHook(() => useGleekGame(), { wrapper: createWrapper() });
    // handleConfigChange は select の値をそのまま受けるので文字列。
    act(() => result.current.handleConfigChange('cpuDifficulty', '2'));
    act(() => result.current.reset());
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', { config: { ...DEFAULT_GLEEK_CONFIG, cpuDifficulty: 2 } }),
    );
  });
});
