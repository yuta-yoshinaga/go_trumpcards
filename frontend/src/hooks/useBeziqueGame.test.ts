import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { createElement } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { beziqueApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { makeBeziqueState } from '../test/stateFactories';
import { DEFAULT_BEZIQUE_CONFIG, useBeziqueGame } from './useBeziqueGame';

vi.mock('../api/gameApi', () => ({
  beziqueApi: { exec: vi.fn() },
  actionLogApi: { bezique: vi.fn() },
}));

const mockExec = vi.mocked(beziqueApi.exec);

function createWrapper() {
  const queryClient = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  return ({ children }: { children: ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children);
}

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(makeBeziqueState());
});

describe('useBeziqueGame', () => {
  it('reset dispatches reset with the default config', async () => {
    const { result } = renderHook(() => useBeziqueGame(), { wrapper: createWrapper() });
    act(() => result.current.reset());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', { config: DEFAULT_BEZIQUE_CONFIG }));
  });

  it('handlePlay does nothing without exactly one selected card', async () => {
    const { result } = renderHook(() => useBeziqueGame(), { wrapper: createWrapper() });
    act(() => result.current.handlePlay());
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('handlePlay dispatches play for the single selected card', async () => {
    const { result } = renderHook(() => useBeziqueGame(), { wrapper: createWrapper() });
    act(() => result.current.toggleCard(2));
    act(() => result.current.handlePlay());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { cardIndex: 2 }));
  });

  it('handleMeld dispatches the given meld index', async () => {
    const { result } = renderHook(() => useBeziqueGame(), { wrapper: createWrapper() });
    act(() => result.current.handleMeld(1));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('meld', { meldIndex: 1 }));
  });

  it('handleSkipMeld dispatches skip', async () => {
    const { result } = renderHook(() => useBeziqueGame(), { wrapper: createWrapper() });
    act(() => result.current.handleSkipMeld());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('skip'));
  });

  it('handleNextRound dispatches next', async () => {
    const { result } = renderHook(() => useBeziqueGame(), { wrapper: createWrapper() });
    act(() => result.current.handleNextRound());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
  });

  it('handleConfigChange updates the config used by reset', async () => {
    const { result } = renderHook(() => useBeziqueGame(), { wrapper: createWrapper() });
    act(() => result.current.handleConfigChange('targetScore', '500'));
    act(() => result.current.reset());
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', { config: { ...DEFAULT_BEZIQUE_CONFIG, targetScore: 500 } }),
    );
  });
});
