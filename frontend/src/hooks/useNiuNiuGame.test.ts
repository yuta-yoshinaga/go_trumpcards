import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import { createElement, type ReactNode } from 'react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { niuniuApi } from '../api/gameApi';
import type { NiuNiuResponse } from '../types/card';
import { useNiuNiuGame } from './useNiuNiuGame';

function makeWrapper() {
  const client = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  return ({ children }: { children: ReactNode }) => createElement(QueryClientProvider, { client }, children);
}

vi.mock('../api/gameApi', () => ({
  niuniuApi: { exec: vi.fn() },
  actionLogApi: { niuniu: vi.fn() },
}));

const mockExec = vi.mocked(niuniuApi.exec);

const baseState: NiuNiuResponse = {
  seats: [
    { name: 'あなた', isCpu: false },
    { name: 'CPU1', isCpu: true },
    { name: 'CPU2', isCpu: true },
    { name: '親', isCpu: true },
  ],
  bankerIdx: 3,
  chips: 1000,
  maxMultiplier: 3,
  bankerRankKey: '',
  phase: 1,
  message: '',
};

describe('useNiuNiuGame', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockExec.mockResolvedValue(baseState);
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it('calls reset on mount', async () => {
    renderHook(() => useNiuNiuGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  it('forwards reset', async () => {
    const { result } = renderHook(() => useNiuNiuGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    mockExec.mockClear();

    act(() => result.current.handleReset());
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset'));
  });

  // Betting is the only action -- it deals and settles in one call.
  it('sends the stake with the bet', async () => {
    const { result } = renderHook(() => useNiuNiuGame(), { wrapper: makeWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalled());
    mockExec.mockClear();

    act(() => result.current.handleBet(100));
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bet', 100));
  });
});
