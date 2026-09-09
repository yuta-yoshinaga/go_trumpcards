import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import { createElement, type ReactNode } from 'react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { quinzeApi } from '../api/gameApi';
import type { QuinzeResponse } from '../types/card';
import { useQuinzeGame } from './useQuinzeGame';

function makeWrapper() {
  const client = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  return ({ children }: { children: ReactNode }) => createElement(QueryClientProvider, { client }, children);
}

vi.mock('../api/gameApi', () => ({
  quinzeApi: { exec: vi.fn() },
  actionLogApi: { quinze: vi.fn() },
}));

const mockExec = vi.mocked(quinzeApi.exec);

const baseState: QuinzeResponse = {
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
  targetPoints: 15,
  cpuStandPoints: 11,
  canHit: false,
  canStand: false,

  message: '',
};

describe('useQuinzeGame', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockExec.mockResolvedValue(baseState);
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it('calls reset on mount', async () => {
    renderHook(() => useQuinzeGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('forwards the commands that carry no amount', async () => {
    const { result } = renderHook(() => useQuinzeGame(), { wrapper: makeWrapper() });
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
});
