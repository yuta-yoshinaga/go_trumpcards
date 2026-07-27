import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { createElement } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { ultiApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { makeUltiState } from '../test/stateFactories';
import { DEFAULT_ULTI_CONFIG, useUltiGame } from './useUltiGame';

vi.mock('../api/gameApi', () => ({
  ultiApi: { exec: vi.fn() },
  actionLogApi: { ulti: vi.fn() },
}));

const mockExec = vi.mocked(ultiApi.exec);

function createWrapper() {
  const queryClient = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  return ({ children }: { children: ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children);
}

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(makeUltiState());
});

describe('useUltiGame', () => {
  it('reset dispatches reset with the default config', async () => {
    const { result } = renderHook(() => useUltiGame(), { wrapper: createWrapper() });
    act(() => result.current.reset());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', { config: DEFAULT_ULTI_CONFIG }));
  });

  it('handleBid dispatches a betli contract with no trump', async () => {
    const { result } = renderHook(() => useUltiGame(), { wrapper: createWrapper() });
    act(() => result.current.handleBid('betli'));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', { contract: 'betli', trumpSuit: undefined }));
  });

  it('handleBid dispatches a party contract with the chosen trump suit', async () => {
    const { result } = renderHook(() => useUltiGame(), { wrapper: createWrapper() });
    act(() => result.current.handleBid('party', 3));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', { contract: 'party', trumpSuit: 3 }));
  });

  it('handleDiscard does nothing without exactly two selected cards', async () => {
    const { result } = renderHook(() => useUltiGame(), { wrapper: createWrapper() });
    act(() => result.current.toggleCard(0));
    act(() => result.current.handleDiscard());
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('handleDiscard dispatches discard with the two selected card indices', async () => {
    const { result } = renderHook(() => useUltiGame(), { wrapper: createWrapper() });
    act(() => result.current.toggleCard(0));
    act(() => result.current.toggleCard(3));
    act(() => result.current.handleDiscard());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('discard', { cardIndices: [0, 3] }));
  });

  it('handlePlay does nothing without exactly one selected card', async () => {
    const { result } = renderHook(() => useUltiGame(), { wrapper: createWrapper() });
    act(() => result.current.handlePlay());
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('handlePlay dispatches play for the single selected card', async () => {
    const { result } = renderHook(() => useUltiGame(), { wrapper: createWrapper() });
    act(() => result.current.toggleCard(2));
    act(() => result.current.handlePlay());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { cardIndex: 2 }));
  });

  it('handleNextTrick and handleNextRound dispatch their commands', async () => {
    const { result } = renderHook(() => useUltiGame(), { wrapper: createWrapper() });
    act(() => result.current.handleNextTrick());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
    act(() => result.current.handleNextRound());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
  });

  it('handleConfigChange updates the config used by reset', async () => {
    const { result } = renderHook(() => useUltiGame(), { wrapper: createWrapper() });
    act(() => result.current.handleConfigChange('targetRounds', '7'));
    act(() => result.current.reset());
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', { config: { ...DEFAULT_ULTI_CONFIG, targetRounds: 7 } }),
    );
  });
});
