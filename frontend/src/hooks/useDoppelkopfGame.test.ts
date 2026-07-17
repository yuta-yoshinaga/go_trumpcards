import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { createElement } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { doppelkopfApi } from '../api/gameApi';
import { makeDoppelkopfState } from '../test/stateFactories';
import { DEFAULT_DOPPELKOPF_CONFIG, useDoppelkopfGame } from './useDoppelkopfGame';

vi.mock('../api/gameApi', () => ({
  doppelkopfApi: { exec: vi.fn() },
}));

const mockExec = vi.mocked(doppelkopfApi.exec);

function createWrapper() {
  const queryClient = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  return ({ children }: { children: ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children);
}

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(makeDoppelkopfState());
});

describe('useDoppelkopfGame', () => {
  it('reset dispatches the reset command with the default config', async () => {
    const { result } = renderHook(() => useDoppelkopfGame(), { wrapper: createWrapper() });
    await act(async () => {
      result.current.reset();
    });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', { config: DEFAULT_DOPPELKOPF_CONFIG }));
  });

  it('handlePlay dispatches only when exactly one card is selected', async () => {
    const { result } = renderHook(() => useDoppelkopfGame(), { wrapper: createWrapper() });
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

  it('handleAnnounce dispatches the announce command', async () => {
    const { result } = renderHook(() => useDoppelkopfGame(), { wrapper: createWrapper() });
    await act(async () => {
      result.current.handleAnnounce();
    });
    expect(mockExec).toHaveBeenCalledWith('announce');
  });

  it('handleHint dispatches the hint command and toggles hintLoading', async () => {
    const { result } = renderHook(() => useDoppelkopfGame(), { wrapper: createWrapper() });
    await act(async () => {
      await result.current.handleHint();
    });
    expect(mockExec).toHaveBeenCalledWith('hint');
    expect(result.current.hintLoading).toBe(false);
  });

  it('handleNextTrick / handleNextRound dispatch their commands', async () => {
    const { result } = renderHook(() => useDoppelkopfGame(), { wrapper: createWrapper() });
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
    const { result } = renderHook(() => useDoppelkopfGame(), { wrapper: createWrapper() });
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
