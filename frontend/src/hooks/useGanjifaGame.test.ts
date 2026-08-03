import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { createElement } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { ganjifaApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { makeGanjifaState } from '../test/stateFactories';
import { DEFAULT_GANJIFA_CONFIG, useGanjifaGame } from './useGanjifaGame';

function createWrapper() {
  const queryClient = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  return ({ children }: { children: ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children);
}

vi.mock('../api/gameApi', () => ({
  ganjifaApi: { exec: vi.fn() },
}));

const mockExec = vi.mocked(ganjifaApi.exec);

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(makeGanjifaState());
});

describe('useGanjifaGame', () => {
  it('resets with the current config', async () => {
    const { result } = renderHook(() => useGanjifaGame(), { wrapper: createWrapper() });
    act(() => result.current.reset());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', { config: DEFAULT_GANJIFA_CONFIG }));
  });

  it('carries a changed config into the next reset', async () => {
    const { result } = renderHook(() => useGanjifaGame(), { wrapper: createWrapper() });
    act(() => result.current.handleConfigChange('targetRounds', '6'));
    act(() => result.current.reset());
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', { config: { ...DEFAULT_GANJIFA_CONFIG, targetRounds: 6 } }),
    );
  });

  it('plays exactly one selected card', async () => {
    const { result } = renderHook(() => useGanjifaGame(), { wrapper: createWrapper() });
    act(() => result.current.toggleCard(2));
    act(() => result.current.handlePlay());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { cardIndex: 2 }));
  });

  it('does not play with nothing selected', async () => {
    const { result } = renderHook(() => useGanjifaGame(), { wrapper: createWrapper() });
    act(() => result.current.handlePlay());
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });

  // Two selected cards is an ambiguous request, and picking the first would play
  // a card the player did not choose.
  it('does not play with two cards selected', async () => {
    const { result } = renderHook(() => useGanjifaGame(), { wrapper: createWrapper() });
    act(() => result.current.toggleCard(0));
    act(() => result.current.toggleCard(1));
    act(() => result.current.handlePlay());
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('advances the trick and the round', async () => {
    const { result } = renderHook(() => useGanjifaGame(), { wrapper: createWrapper() });
    act(() => result.current.handleNextTrick());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
    act(() => result.current.handleNextRound());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
  });

  it('clears the selection after a successful call', async () => {
    const { result } = renderHook(() => useGanjifaGame(), { wrapper: createWrapper() });
    act(() => result.current.toggleCard(1));
    expect(result.current.selectedCardIndices).toEqual([1]);
    act(() => result.current.handlePlay());
    await waitFor(() => expect(result.current.selectedCardIndices).toEqual([]));
  });
});
