import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { createElement } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { musApi } from '../api/gameApi';
import { makeMusState } from '../test/stateFactories';
import { DEFAULT_MUS_CONFIG, useMusGame } from './useMusGame';

vi.mock('../api/gameApi', () => ({
  musApi: { exec: vi.fn() },
}));

const mockExec = vi.mocked(musApi.exec);

function createWrapper() {
  const queryClient = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  return ({ children }: { children: ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children);
}

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(makeMusState());
});

describe('useMusGame', () => {
  it('reset dispatches the reset command with the default config', async () => {
    const { result } = renderHook(() => useMusGame(), { wrapper: createWrapper() });
    await act(async () => {
      result.current.reset();
    });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', { config: DEFAULT_MUS_CONFIG }));
  });

  it('handleMus dispatches the mus command with the boolean flag', async () => {
    const { result } = renderHook(() => useMusGame(), { wrapper: createWrapper() });
    await act(async () => {
      result.current.handleMus(true);
    });
    expect(mockExec).toHaveBeenCalledWith('mus', { mus: true });
    await act(async () => {
      result.current.handleMus(false);
    });
    expect(mockExec).toHaveBeenCalledWith('mus', { mus: false });
  });

  it('handleDiscard dispatches the selected indices (empty keeps all)', async () => {
    const { result } = renderHook(() => useMusGame(), { wrapper: createWrapper() });
    await act(async () => {
      result.current.handleDiscard();
    });
    expect(mockExec).toHaveBeenCalledWith('discard', { discardIndices: [] });
    await act(async () => {
      result.current.toggleCard(0);
    });
    await act(async () => {
      result.current.toggleCard(2);
    });
    await act(async () => {
      result.current.handleDiscard();
    });
    expect(mockExec).toHaveBeenCalledWith('discard', { discardIndices: [0, 2] });
  });

  it('handleBet dispatches the action and amount', async () => {
    const { result } = renderHook(() => useMusGame(), { wrapper: createWrapper() });
    await act(async () => {
      result.current.handleBet(0);
    });
    expect(mockExec).toHaveBeenCalledWith('bet', { betAction: 0, betAmount: 0 });
    await act(async () => {
      result.current.handleBet(1, 3);
    });
    expect(mockExec).toHaveBeenCalledWith('bet', { betAction: 1, betAmount: 3 });
  });

  it('handleNextRound dispatches the next command', async () => {
    const { result } = renderHook(() => useMusGame(), { wrapper: createWrapper() });
    await act(async () => {
      result.current.handleNextRound();
    });
    expect(mockExec).toHaveBeenCalledWith('next');
  });

  it('clears card selection after a successful action', async () => {
    const { result } = renderHook(() => useMusGame(), { wrapper: createWrapper() });
    await act(async () => {
      result.current.toggleCard(1);
    });
    expect(result.current.selectedCardIndices).toEqual([1]);
    await act(async () => {
      result.current.handleDiscard();
    });
    await waitFor(() => expect(result.current.selectedCardIndices).toEqual([]));
  });
});
