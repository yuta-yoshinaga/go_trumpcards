import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { createElement } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { tarocchiniApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import { makeTarocchiniState } from '../test/stateFactories';
import { DEFAULT_TAROCCHINI_CONFIG, TARGET_ROUNDS_OPTIONS, useTarocchiniGame } from './useTarocchiniGame';

vi.mock('../api/gameApi', () => ({
  tarocchiniApi: { exec: vi.fn() },
  actionLogApi: { tarocchini: vi.fn() },
}));

const mockExec = vi.mocked(tarocchiniApi.exec);

function createWrapper() {
  const queryClient = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  return ({ children }: { children: ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children);
}

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(makeTarocchiniState());
});

describe('useTarocchiniGame', () => {
  it('resets with the current config', async () => {
    const { result } = renderHook(() => useTarocchiniGame(), { wrapper: createWrapper() });
    act(() => result.current.reset());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', { config: DEFAULT_TAROCCHINI_CONFIG }));
  });

  it('carries a changed config into the next reset', async () => {
    const { result } = renderHook(() => useTarocchiniGame(), { wrapper: createWrapper() });
    act(() => result.current.handleConfigChange('targetRounds', '8'));
    act(() => result.current.reset());
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', { config: { ...DEFAULT_TAROCCHINI_CONFIG, targetRounds: 8 } }),
    );
  });

  // ディーラーは 1 局ごとに回る。倍数でない局数はバックエンドが弾くので、
  // 選択肢に出してはならない。
  it('offers only round counts that are a multiple of the player count', () => {
    for (const rounds of TARGET_ROUNDS_OPTIONS) {
      expect(rounds % 4).toBe(0);
    }
  });

  it('buries exactly two cards', async () => {
    const { result } = renderHook(() => useTarocchiniGame(), { wrapper: createWrapper() });
    act(() => result.current.toggleCard(0));
    act(() => result.current.toggleCard(3));
    act(() => result.current.handleScarto());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('scarto', { cardIndices: [0, 3] }));
  });

  it('does not bury with the wrong number selected', async () => {
    const { result } = renderHook(() => useTarocchiniGame(), { wrapper: createWrapper() });
    act(() => result.current.handleScarto());
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();

    act(() => result.current.toggleCard(0));
    act(() => result.current.handleScarto());
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();

    act(() => result.current.toggleCard(1));
    act(() => result.current.toggleCard(2));
    act(() => result.current.handleScarto());
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('plays exactly one selected card', async () => {
    const { result } = renderHook(() => useTarocchiniGame(), { wrapper: createWrapper() });
    act(() => result.current.toggleCard(2));
    act(() => result.current.handlePlay());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', { cardIndex: 2 }));
  });

  it('does not play with nothing or with two cards selected', async () => {
    const { result } = renderHook(() => useTarocchiniGame(), { wrapper: createWrapper() });
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
    const { result } = renderHook(() => useTarocchiniGame(), { wrapper: createWrapper() });
    act(() => result.current.handleNextTrick());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
    act(() => result.current.handleNextRound());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
  });

  it('clears the selection after a successful call', async () => {
    const { result } = renderHook(() => useTarocchiniGame(), { wrapper: createWrapper() });
    act(() => result.current.toggleCard(1));
    expect(result.current.selectedCardIndices).toEqual([1]);
    act(() => result.current.handlePlay());
    await waitFor(() => expect(result.current.selectedCardIndices).toEqual([]));
  });
});
