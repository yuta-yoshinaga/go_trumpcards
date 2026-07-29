import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import { createElement, type ReactNode } from 'react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { settemezzoApi } from '../api/gameApi';
import type { SetteEMezzoResponse } from '../types/card';
import { useSetteEMezzoGame } from './useSetteEMezzoGame';

function makeWrapper() {
  const client = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  return ({ children }: { children: ReactNode }) => createElement(QueryClientProvider, { client }, children);
}

vi.mock('../api/gameApi', () => ({
  settemezzoApi: { exec: vi.fn() },
  actionLogApi: { settemezzo: vi.fn() },
}));

const mockExec = vi.mocked(settemezzoApi.exec);

const baseState: SetteEMezzoResponse = {
  seats: [
    { name: 'あなた', isCpu: false },
    { name: 'CPU1', isCpu: true },
    { name: 'CPU2', isCpu: true },
  ],
  bankerIdx: 1,
  isHumanBanker: false,
  chips: 1000,
  activeSeat: 0,
  nextBanker: -1,
  lastResult: '',
  phase: 1,
  targetHalves: 15,
  canHit: false,
  canStand: false,
  canSetMatta: false,
  message: '',
};

describe('useSetteEMezzoGame', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockExec.mockResolvedValue(baseState);
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it('calls reset on mount', async () => {
    renderHook(() => useSetteEMezzoGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('forwards the commands that carry no amount', async () => {
    const { result } = renderHook(() => useSetteEMezzoGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    mockExec.mockClear();

    act(() => result.current.handleReset());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    act(() => result.current.handleDeal());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('deal'));

    act(() => result.current.handleHit());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('hit'));

    act(() => result.current.handleStand());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('stand'));

    act(() => result.current.handleBankerHit());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bankerhit'));

    act(() => result.current.handleBankerStand());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bankerstand'));
  });

  // The matta travels in HALVES end to end. Sending points here would halve
  // every value the domain sees.
  it('sends the stake and the matta in the amount slot', async () => {
    const { result } = renderHook(() => useSetteEMezzoGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    mockExec.mockClear();

    act(() => result.current.handleBet(100));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bet', 100));

    act(() => result.current.handleMatta(6));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('matta', 6));
  });
});
