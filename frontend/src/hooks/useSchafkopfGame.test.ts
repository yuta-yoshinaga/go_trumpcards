import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { createElement } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { schafkopfApi } from '../api/gameApi';
import { makeSchafkopfState } from '../test/stateFactories';
import { DEFAULT_SCHAFKOPF_CONFIG, useSchafkopfGame } from './useSchafkopfGame';

vi.mock('../api/gameApi', () => ({
  schafkopfApi: { exec: vi.fn() },
}));

const mockExec = vi.mocked(schafkopfApi.exec);

function createWrapper() {
  const queryClient = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  return ({ children }: { children: ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children);
}

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(makeSchafkopfState());
});

describe('useSchafkopfGame', () => {
  it('reset dispatches the reset command with the default config', async () => {
    const { result } = renderHook(() => useSchafkopfGame(), { wrapper: createWrapper() });
    await act(async () => {
      result.current.reset();
    });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', { config: DEFAULT_SCHAFKOPF_CONFIG }));
  });

  it('handlePick / handlePass dispatch the pick command with the right flag', async () => {
    const { result } = renderHook(() => useSchafkopfGame(), { wrapper: createWrapper() });
    await act(async () => {
      result.current.handlePick();
    });
    expect(mockExec).toHaveBeenCalledWith('pick', { pick: true, contract: 0, soloSuit: undefined });
    await act(async () => {
      result.current.handlePass();
    });
    expect(mockExec).toHaveBeenCalledWith('pick', { pick: false });
  });

  it('handleDeclare sends the contract, and a Solo carries its trump suit', async () => {
    const { result } = renderHook(() => useSchafkopfGame(), { wrapper: createWrapper() });
    await act(async () => {
      result.current.handleDeclare(1);
    });
    expect(mockExec).toHaveBeenCalledWith('pick', { pick: true, contract: 1, soloSuit: undefined });
    // **Solo must carry its suit.** Dropping it here would leave the backend
    // to guess, and a Solo without a named trump is not a legal declaration.
    await act(async () => {
      result.current.handleDeclare(2, 3);
    });
    expect(mockExec).toHaveBeenCalledWith('pick', { pick: true, contract: 2, soloSuit: 3 });
  });

  it('handleCall dispatches the call command with the suit', async () => {
    const { result } = renderHook(() => useSchafkopfGame(), { wrapper: createWrapper() });
    await act(async () => {
      result.current.handleCall(2);
    });
    expect(mockExec).toHaveBeenCalledWith('call', { callSuit: 2 });
  });

  it('handlePlay dispatches only when exactly one card is selected', async () => {
    const { result } = renderHook(() => useSchafkopfGame(), { wrapper: createWrapper() });
    await act(async () => {
      result.current.handlePlay();
    });
    expect(mockExec).not.toHaveBeenCalledWith('play', expect.anything());
    await act(async () => {
      result.current.toggleCard(1);
    });
    await act(async () => {
      result.current.handlePlay();
    });
    expect(mockExec).toHaveBeenCalledWith('play', { cardIndex: 1 });
  });

  it('handleNextTrick / handleNextRound dispatch their commands', async () => {
    const { result } = renderHook(() => useSchafkopfGame(), { wrapper: createWrapper() });
    await act(async () => {
      result.current.handleNextTrick();
    });
    expect(mockExec).toHaveBeenCalledWith('next');
    await act(async () => {
      result.current.handleNextRound();
    });
    expect(mockExec).toHaveBeenCalledWith('nextround');
  });

  it('clears card selection after a successful action', async () => {
    const { result } = renderHook(() => useSchafkopfGame(), { wrapper: createWrapper() });
    await act(async () => {
      result.current.toggleCard(1);
    });
    expect(result.current.selectedCardIndices).toEqual([1]);
    await act(async () => {
      result.current.handlePlay();
    });
    await waitFor(() => expect(result.current.selectedCardIndices).toEqual([]));
  });
});
