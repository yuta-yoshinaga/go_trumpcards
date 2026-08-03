import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { createElement } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { aluetteApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { makeAluetteState } from '../test/stateFactories';
import { ALUETTE_HAND_SIZE } from '../types/card';
import { DEFAULT_ALUETTE_CONFIG, TARGET_POINTS_OPTIONS, useAluetteGame } from './useAluetteGame';

vi.mock('../api/gameApi', () => ({
  aluetteApi: { exec: vi.fn() },
  actionLogApi: { aluette: vi.fn() },
}));

const mockExec = vi.mocked(aluetteApi.exec);

function createWrapper() {
  const queryClient = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  return ({ children }: { children: ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children);
}

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(makeAluetteState());
});

describe('useAluetteGame', () => {
  it('resets with the current config', async () => {
    const { result } = renderHook(() => useAluetteGame(), { wrapper: createWrapper() });
    act(() => result.current.reset());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', { config: DEFAULT_ALUETTE_CONFIG }));
  });

  it('carries a changed config into the next reset', async () => {
    const { result } = renderHook(() => useAluetteGame(), { wrapper: createWrapper() });
    act(() => result.current.handleConfigChange('targetPoints', '7'));
    act(() => result.current.reset());
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', { config: { ...DEFAULT_ALUETTE_CONFIG, targetPoints: 7 } }),
    );
  });

  // メーヌは 3 勝で 1 点。到達不能な目標や 0 点は出してはならない。
  it('offers only positive target scores', () => {
    for (const points of TARGET_POINTS_OPTIONS) {
      expect(points).toBeGreaterThan(0);
    }
  });

  it('plays exactly one selected card', async () => {
    const { result } = renderHook(() => useAluetteGame(), { wrapper: createWrapper() });
    act(() => result.current.toggleCard(2));
    act(() => result.current.handlePlay());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { cardIndex: 2 }));
  });

  it('does not play with nothing or with two cards selected', async () => {
    const { result } = renderHook(() => useAluetteGame(), { wrapper: createWrapper() });
    act(() => result.current.handlePlay());
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();

    act(() => result.current.toggleCard(0));
    act(() => result.current.toggleCard(1));
    act(() => result.current.handlePlay());
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });

  // **捨て札はこのゲームに無い。**タロー系から写した痕跡が残っていないこと。
  it('exposes no discard action', () => {
    const { result } = renderHook(() => useAluetteGame(), { wrapper: createWrapper() });
    expect(result.current).not.toHaveProperty('handleScarto');
    expect(result.current).not.toHaveProperty('handleDiscard');
  });

  // 手札は 5 枚しか無いので、それを超える位置を選べてしまうと index が破綻する。
  it('selects within the hand', async () => {
    const { result } = renderHook(() => useAluetteGame(), { wrapper: createWrapper() });
    act(() => result.current.toggleCard(ALUETTE_HAND_SIZE - 1));
    act(() => result.current.handlePlay());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { cardIndex: ALUETTE_HAND_SIZE - 1 }));
  });

  it('advances the trick and the mene', async () => {
    const { result } = renderHook(() => useAluetteGame(), { wrapper: createWrapper() });
    act(() => result.current.handleNextTrick());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
    act(() => result.current.handleNextRound());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
  });

  it('clears the selection after a successful call', async () => {
    const { result } = renderHook(() => useAluetteGame(), { wrapper: createWrapper() });
    act(() => result.current.toggleCard(1));
    expect(result.current.selectedCardIndices).toEqual([1]);
    act(() => result.current.handlePlay());
    await waitFor(() => expect(result.current.selectedCardIndices).toEqual([]));
  });
});
