import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { createElement } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { sheepsheadApi } from '../api/gameApi';
import { makeSheepsheadState } from '../test/stateFactories';
import { DEFAULT_SHEEPSHEAD_CONFIG, useSheepsheadGame } from './useSheepsheadGame';

vi.mock('../api/gameApi', () => ({
  sheepsheadApi: { exec: vi.fn() },
}));

const mockExec = vi.mocked(sheepsheadApi.exec);

function createWrapper() {
  const queryClient = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  return ({ children }: { children: ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children);
}

beforeEach(() => {
  mockExec.mockReset();
  mockExec.mockResolvedValue(makeSheepsheadState());
});

describe('useSheepsheadGame', () => {
  it('reset dispatches the reset command with the default config', async () => {
    const { result } = renderHook(() => useSheepsheadGame(), { wrapper: createWrapper() });
    await act(async () => {
      result.current.reset();
    });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', { config: DEFAULT_SHEEPSHEAD_CONFIG }));
  });

  it('handlePick / handlePass dispatch the pick command with the right flag', async () => {
    const { result } = renderHook(() => useSheepsheadGame(), { wrapper: createWrapper() });
    await act(async () => {
      result.current.handlePick();
    });
    expect(mockExec).toHaveBeenCalledWith('pick', { pick: true });
    await act(async () => {
      result.current.handlePass();
    });
    expect(mockExec).toHaveBeenCalledWith('pick', { pick: false });
  });

  it('handleBury dispatches only when exactly two cards are selected', async () => {
    const { result } = renderHook(() => useSheepsheadGame(), { wrapper: createWrapper() });
    // One card selected — no dispatch.
    await act(async () => {
      result.current.toggleCard(0);
    });
    await act(async () => {
      result.current.handleBury();
    });
    expect(mockExec).not.toHaveBeenCalledWith('bury', expect.anything());
    // Two cards selected — dispatches.
    await act(async () => {
      result.current.toggleCard(1);
    });
    await act(async () => {
      result.current.handleBury();
    });
    expect(mockExec).toHaveBeenCalledWith('bury', { buryIndices: [0, 1] });
  });

  it('handleCall dispatches the call command with the suit', async () => {
    const { result } = renderHook(() => useSheepsheadGame(), { wrapper: createWrapper() });
    await act(async () => {
      result.current.handleCall(2);
    });
    expect(mockExec).toHaveBeenCalledWith('call', { callSuit: 2 });
  });

  it('handlePlay dispatches only when exactly one card is selected', async () => {
    const { result } = renderHook(() => useSheepsheadGame(), { wrapper: createWrapper() });
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
    const { result } = renderHook(() => useSheepsheadGame(), { wrapper: createWrapper() });
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
    const { result } = renderHook(() => useSheepsheadGame(), { wrapper: createWrapper() });
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
