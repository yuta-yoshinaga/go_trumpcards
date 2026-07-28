import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { createElement } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { ecarteApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { makeEcarteState } from '../test/stateFactories';
import { DEFAULT_ECARTE_CONFIG, useEcarteGame } from './useEcarteGame';

vi.mock('../api/gameApi', () => ({
  ecarteApi: { exec: vi.fn() },
  actionLogApi: { ecarte: vi.fn() },
}));

const mockExec = vi.mocked(ecarteApi.exec);

function createWrapper() {
  const queryClient = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  return ({ children }: { children: ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children);
}

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(makeEcarteState());
});

describe('useEcarteGame', () => {
  it('reset dispatches reset with the default config', async () => {
    const { result } = renderHook(() => useEcarteGame(), { wrapper: createWrapper() });
    act(() => result.current.reset());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', { config: DEFAULT_ECARTE_CONFIG }));
  });

  it('handlePlay does nothing without exactly one selected card', async () => {
    const { result } = renderHook(() => useEcarteGame(), { wrapper: createWrapper() });
    act(() => result.current.handlePlay());
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('handlePlay dispatches play for the single selected card', async () => {
    const { result } = renderHook(() => useEcarteGame(), { wrapper: createWrapper() });
    act(() => result.current.toggleCard(2));
    act(() => result.current.handlePlay());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { cardIndex: 2 }));
  });

  it('handlePropose dispatches propose', async () => {
    const { result } = renderHook(() => useEcarteGame(), { wrapper: createWrapper() });
    act(() => result.current.handlePropose());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('propose'));
  });

  it('handleStand dispatches stand', async () => {
    const { result } = renderHook(() => useEcarteGame(), { wrapper: createWrapper() });
    act(() => result.current.handleStand());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('stand'));
  });

  it('handleRespond dispatches respond with the accept flag', async () => {
    const { result } = renderHook(() => useEcarteGame(), { wrapper: createWrapper() });
    act(() => result.current.handleRespond(true));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('respond', { accept: true }));
  });

  it('handleDiscard dispatches discard with the selected indices', async () => {
    const { result } = renderHook(() => useEcarteGame(), { wrapper: createWrapper() });
    act(() => result.current.toggleCard(0));
    act(() => result.current.toggleCard(2));
    act(() => result.current.handleDiscard());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('discard', { discardIndices: [0, 2] }));
  });

  it('handleNextRound dispatches next', async () => {
    const { result } = renderHook(() => useEcarteGame(), { wrapper: createWrapper() });
    act(() => result.current.handleNextRound());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
  });

  it('handleConfigChange updates the config used by reset', async () => {
    const { result } = renderHook(() => useEcarteGame(), { wrapper: createWrapper() });
    act(() => result.current.handleConfigChange('targetScore', '7'));
    act(() => result.current.reset());
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', { config: { ...DEFAULT_ECARTE_CONFIG, targetScore: 7 } }),
    );
  });
});
