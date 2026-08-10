import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { createElement } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { spadesApi } from '../api/gameApi';
import type { SpadesResponse } from '../types/card';
import { useSpadesGame } from './useSpadesGame';

vi.mock('../api/gameApi', () => ({
  spadesApi: { exec: vi.fn() },
}));

const mockExec = vi.mocked(spadesApi.exec);

function createWrapper() {
  const queryClient = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  return ({ children }: { children: ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children);
}

const defaultState: SpadesResponse = {
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 13,
      cards: [],
      bid: -1,
      roundScore: 0,
      cumulativeScore: 0,
      trickCount: 0,
      bags: 0,
    },
    {
      id: 1,
      isHuman: false,
      cardCount: 13,
      cards: [],
      bid: -1,
      roundScore: 0,
      cumulativeScore: 0,
      trickCount: 0,
      bags: 0,
    },
    {
      id: 2,
      isHuman: false,
      cardCount: 13,
      cards: [],
      bid: -1,
      roundScore: 0,
      cumulativeScore: 0,
      trickCount: 0,
      bags: 0,
    },
    {
      id: 3,
      isHuman: false,
      cardCount: 13,
      cards: [],
      bid: -1,
      roundScore: 0,
      cumulativeScore: 0,
      trickCount: 0,
      bags: 0,
    },
  ],
  phase: 0,
  roundNumber: 1,
  trickNumber: 1,
  currentPlayerIdx: 0,
  bidPlayerIdx: 0,
  currentTrick: [],
  spadesBroken: false,
  gameEndFlag: false,
  winnerIdx: -1,
  leadPlayerIdx: 0,
  validPlayIndices: [],
  message: '',
  config: { cpuDifficulty: 1, pointLimit: 500, nilBonus: 100, bagPenaltyThreshold: 10 },
};

beforeEach(() => {
  mockExec.mockResolvedValue(defaultState);
});

describe('useSpadesGame', () => {
  it('calls reset on mount', async () => {
    renderHook(() => useSpadesGame(), { wrapper: createWrapper() });
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, {
        cpuDifficulty: 1,
        pointLimit: 500,
        nilBonus: 100,
        bagPenaltyThreshold: 10,
      }),
    );
  });

  it('handleBid dispatches bid command with bid number', async () => {
    const { result } = renderHook(() => useSpadesGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handleBid(3);
    });

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', 3));
  });

  it('handleBid dispatches nil bid (0)', async () => {
    const { result } = renderHook(() => useSpadesGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handleBid(0);
    });

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('bid', 0));
  });
});
