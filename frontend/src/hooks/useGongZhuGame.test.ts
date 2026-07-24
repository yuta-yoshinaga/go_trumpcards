import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { createElement } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { gongzhuApi } from '../api/gameApi';
import type { GongZhuResponse } from '../types/card';
import { useGongZhuGame } from './useGongZhuGame';

vi.mock('../api/gameApi', () => ({
  gongzhuApi: { exec: vi.fn() },
}));

const mockExec = vi.mocked(gongzhuApi.exec);

function createWrapper() {
  const queryClient = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  return ({ children }: { children: ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children);
}

const defaultState: GongZhuResponse = {
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 13,
      cards: [],
      capturedPointCards: [],
      roundScore: 0,
      cumulativeScore: 0,
      trickCount: 0,
    },
    {
      id: 1,
      isHuman: false,
      cardCount: 13,
      cards: [],
      capturedPointCards: [],
      roundScore: 0,
      cumulativeScore: 0,
      trickCount: 0,
    },
    {
      id: 2,
      isHuman: false,
      cardCount: 13,
      cards: [],
      capturedPointCards: [],
      roundScore: 0,
      cumulativeScore: 0,
      trickCount: 0,
    },
    {
      id: 3,
      isHuman: false,
      cardCount: 13,
      cards: [],
      capturedPointCards: [],
      roundScore: 0,
      cumulativeScore: 0,
      trickCount: 0,
    },
  ],
  phase: 0,
  roundNumber: 1,
  trickNumber: 0,
  currentPlayerIdx: 0,
  currentTrick: [],
  heartsBroken: false,
  exposed: { pig: false, sheep: false, ace: false, doubler: false },
  exposableIndices: [],
  gameEndFlag: false,
  winnerIdx: -1,
  leadPlayerIdx: 0,
  message: '',
  config: { cpuDifficulty: 1, pointLimit: 1000 },
};

beforeEach(() => {
  mockExec.mockResolvedValue(defaultState);
});

describe('useGongZhuGame', () => {
  it('calls reset on mount', async () => {
    renderHook(() => useGongZhuGame(), { wrapper: createWrapper() });
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, {
        cpuDifficulty: 1,
        pointLimit: 1000,
      }),
    );
  });

  it('returns initial state after mount', async () => {
    const { result } = renderHook(() => useGongZhuGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).toEqual(defaultState));
  });

  it('handleExpose dispatches expose with selected indices', async () => {
    const { result } = renderHook(() => useGongZhuGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    act(() => {
      result.current.toggleCard(0);
      result.current.toggleCard(2);
    });

    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handleExpose();
    });

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('expose', [0, 2]));
  });

  it('handleExpose dispatches expose with empty selection', async () => {
    const { result } = renderHook(() => useGongZhuGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handleExpose();
    });

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('expose', []));
  });

  it('handlePlay dispatches play with single selected card', async () => {
    const { result } = renderHook(() => useGongZhuGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    act(() => {
      result.current.toggleCard(3);
    });

    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handlePlay();
    });

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', undefined, 3));
  });

  it('handleNextTrick dispatches next command', async () => {
    const { result } = renderHook(() => useGongZhuGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handleNextTrick();
    });

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('next'));
  });

  it('handleNextRound dispatches nextround command', async () => {
    const { result } = renderHook(() => useGongZhuGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handleNextRound();
    });

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
  });

  it('handleConfigChange updates config with valid number', async () => {
    const { result } = renderHook(() => useGongZhuGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    act(() => {
      result.current.handleConfigChange('pointLimit', '2000');
    });

    expect(result.current.gongzhuConfig.pointLimit).toBe(2000);
  });

  it('handleConfigChange ignores NaN values', async () => {
    const { result } = renderHook(() => useGongZhuGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    act(() => {
      result.current.handleConfigChange('pointLimit', 'abc');
    });

    expect(result.current.gongzhuConfig.pointLimit).toBe(1000);
  });

  it('clears selection on success', async () => {
    const { result } = renderHook(() => useGongZhuGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    act(() => {
      result.current.toggleCard(0);
    });
    expect(result.current.selectedCardIndices).toEqual([0]);

    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handleExpose();
    });

    await waitFor(() => expect(result.current.selectedCardIndices).toEqual([]));
  });
});
