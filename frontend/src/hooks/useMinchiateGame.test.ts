import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { createElement } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { minchiateApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { makeMinchiateState } from '../test/stateFactories';
import { MINCHIATE_SURPLUS } from '../types/card';
import { DEFAULT_MINCHIATE_CONFIG, TARGET_ROUNDS_OPTIONS, useMinchiateGame } from './useMinchiateGame';

vi.mock('../api/gameApi', () => ({
  minchiateApi: { exec: vi.fn() },
  actionLogApi: { minchiate: vi.fn() },
}));

const mockExec = vi.mocked(minchiateApi.exec);

function createWrapper() {
  const queryClient = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  return ({ children }: { children: ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children);
}

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(makeMinchiateState());
});

describe('useMinchiateGame', () => {
  it('resets with the current config', async () => {
    const { result } = renderHook(() => useMinchiateGame(), { wrapper: createWrapper() });
    act(() => result.current.reset());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', { config: DEFAULT_MINCHIATE_CONFIG }));
  });

  it('carries a changed config into the next reset', async () => {
    const { result } = renderHook(() => useMinchiateGame(), { wrapper: createWrapper() });
    act(() => result.current.handleConfigChange('targetRounds', '8'));
    act(() => result.current.reset());
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', { config: { ...DEFAULT_MINCHIATE_CONFIG, targetRounds: 8 } }),
    );
  });

  // ディーラーは 1 局ごとに回る。倍数でない局数はバックエンドが弾くので、
  // 選択肢に出してはならない。
  it('offers only round counts that are a multiple of the player count', () => {
    for (const rounds of TARGET_ROUNDS_OPTIONS) {
      expect(rounds % 4).toBe(0);
    }
  });

  it('buries exactly the surplus count', async () => {
    const { result } = renderHook(() => useMinchiateGame(), { wrapper: createWrapper() });
    const indices = Array.from({ length: MINCHIATE_SURPLUS }, (_, i) => i);
    for (const i of indices) {
      act(() => result.current.toggleCard(i));
    }
    act(() => result.current.handleScarto());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('scarto', { cardIndices: indices }));
  });

  it('does not bury with the wrong number selected', async () => {
    const { result } = renderHook(() => useMinchiateGame(), { wrapper: createWrapper() });
    act(() => result.current.handleScarto());
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();

    // 1 枚足りない。
    for (let i = 0; i < MINCHIATE_SURPLUS - 1; i++) {
      act(() => result.current.toggleCard(i));
    }
    act(() => result.current.handleScarto());
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();

    // 1 枚多い。
    act(() => result.current.toggleCard(MINCHIATE_SURPLUS - 1));
    act(() => result.current.toggleCard(MINCHIATE_SURPLUS));
    act(() => result.current.handleScarto());
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('plays exactly one selected card', async () => {
    const { result } = renderHook(() => useMinchiateGame(), { wrapper: createWrapper() });
    act(() => result.current.toggleCard(2));
    act(() => result.current.handlePlay());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { cardIndex: 2 }));
  });

  it('does not play with nothing or with two cards selected', async () => {
    const { result } = renderHook(() => useMinchiateGame(), { wrapper: createWrapper() });
    act(() => result.current.handlePlay());
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();

    act(() => result.current.toggleCard(0));
    act(() => result.current.toggleCard(1));
    act(() => result.current.handlePlay());
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('advances the trick and the round', async () => {
    const { result } = renderHook(() => useMinchiateGame(), { wrapper: createWrapper() });
    act(() => result.current.handleNextTrick());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
    act(() => result.current.handleNextRound());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
  });

  it('clears the selection after a successful call', async () => {
    const { result } = renderHook(() => useMinchiateGame(), { wrapper: createWrapper() });
    act(() => result.current.toggleCard(1));
    expect(result.current.selectedCardIndices).toEqual([1]);
    act(() => result.current.handlePlay());
    await waitFor(() => expect(result.current.selectedCardIndices).toEqual([]));
  });
});
