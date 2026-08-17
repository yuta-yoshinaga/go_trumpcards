import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import { createElement, type ReactNode } from 'react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { pontoonApi } from '../api/gameApi';
import type { PontoonResponse } from '../types/card';
import { usePontoonGame } from './usePontoonGame';

function makeWrapper() {
  const client = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  return ({ children }: { children: ReactNode }) => createElement(QueryClientProvider, { client }, children);
}

vi.mock('../api/gameApi', () => ({
  pontoonApi: { exec: vi.fn() },
  actionLogApi: { pontoon: vi.fn() },
}));

const mockExec = vi.mocked(pontoonApi.exec);

const baseState: PontoonResponse = {
  seats: [
    { name: 'あなた', isCpu: false, hands: [] },
    { name: 'CPU1', isCpu: true, hands: [] },
    { name: 'CPU2', isCpu: true, hands: [] },
  ],
  bankerIdx: 1,
  isHumanBanker: false,
  chips: 1000,
  activeSeat: 0,
  activeHand: 0,
  nextBanker: -1,
  lastResult: '',
  phase: 1,
  canStick: false,
  canTwist: false,
  canBuy: false,
  canSplit: false,
  stickMin: 15,
  cpuStickMin: 17,
  message: '',
};

describe('usePontoonGame', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockExec.mockResolvedValue(baseState);
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it('calls reset on mount', async () => {
    renderHook(() => usePontoonGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('forwards the commands that carry no amount', async () => {
    const { result } = renderHook(() => usePontoonGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    mockExec.mockClear();

    act(() => result.current.handleReset());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));

    act(() => result.current.handleDeal());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('deal'));

    act(() => result.current.handleStick());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('stick'));

    act(() => result.current.handleTwist());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('twist'));

    act(() => result.current.handleSplit());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('split'));

    act(() => result.current.handleBankerTwist());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bankertwist'));

    act(() => result.current.handleBankerStay());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bankerstay'));
  });

  // Only bet and buy carry a number, and sending it in the wrong slot means the
  // backend rejects the whole command.
  it('sends the amount for bet and buy', async () => {
    const { result } = renderHook(() => usePontoonGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    mockExec.mockClear();

    act(() => result.current.handleBet(100));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bet', 100));

    act(() => result.current.handleBuy(50));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('buy', 50));
  });
});
