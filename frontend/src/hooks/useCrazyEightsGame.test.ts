import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { createElement } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { crazyeightsApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import type { CrazyEightsResponse } from '../types/card';
import { useCrazyEightsGame } from './useCrazyEightsGame';

vi.mock('../api/gameApi', () => ({
  crazyeightsApi: { exec: vi.fn() },
}));

const mockExec = vi.mocked(crazyeightsApi.exec);

function createWrapper() {
  const queryClient = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  return ({ children }: { children: ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children);
}

const defaultState: CrazyEightsResponse = {
  players: [
    { id: 0, isHuman: true, cardCount: 5, cards: [{ design: 'SPADE', value: 8 }], roundScore: 0, cumulativeScore: 0 },
    { id: 1, isHuman: false, cardCount: 5, cards: [], roundScore: 0, cumulativeScore: 0 },
    { id: 2, isHuman: false, cardCount: 5, cards: [], roundScore: 0, cumulativeScore: 0 },
    { id: 3, isHuman: false, cardCount: 5, cards: [], roundScore: 0, cumulativeScore: 0 },
  ],
  phase: 0,
  roundNumber: 1,
  currentPlayerIdx: 0,
  discardTop: { design: 'HEART', value: 7 },
  drawPileCount: 30,
  chosenSuit: 0,
  gameEndFlag: false,
  winnerIdx: -1,
  message: '',
  config: { cpuDifficulty: 1, pointLimit: 200 },
};

beforeEach(() => {
  mockExec.mockResolvedValue(defaultState);
});

describe('useCrazyEightsGame', () => {
  it('calls reset on mount with default config', async () => {
    renderHook(() => useCrazyEightsGame(), { wrapper: createWrapper() });
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, {
        cpuDifficulty: 1,
        pointLimit: 200,
      }),
    );
  });

  it('returns initial state after mount', async () => {
    const { result } = renderHook(() => useCrazyEightsGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).toEqual(defaultState));
  });

  it('handlePlay dispatches play with single selected card', async () => {
    const { result } = renderHook(() => useCrazyEightsGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    act(() => {
      result.current.toggleCard(2);
    });

    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handlePlay();
    });

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('play', 2));
  });

  it('handlePlay does nothing when no card selected', async () => {
    const { result } = renderHook(() => useCrazyEightsGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    act(() => {
      result.current.handlePlay();
    });

    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('handlePlay does nothing when multiple cards selected', async () => {
    const { result } = renderHook(() => useCrazyEightsGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    act(() => {
      result.current.toggleCard(0);
      result.current.toggleCard(1);
    });

    mockExec.mockClear();
    act(() => {
      result.current.handlePlay();
    });

    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('handleDraw dispatches draw command', async () => {
    const { result } = renderHook(() => useCrazyEightsGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handleDraw();
    });

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('draw'));
  });

  it('handleChooseSuit dispatches suit command with suit value', async () => {
    const { result } = renderHook(() => useCrazyEightsGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handleChooseSuit(3);
    });

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('suit', undefined, 3));
  });

  it('handleNextRound dispatches nextround command', async () => {
    const { result } = renderHook(() => useCrazyEightsGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handleNextRound();
    });

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
  });

  it('handleConfigChange updates config with valid number', async () => {
    const { result } = renderHook(() => useCrazyEightsGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    act(() => {
      result.current.handleConfigChange('pointLimit', '300');
    });

    expect(result.current.crazyEightsConfig.pointLimit).toBe(300);
  });

  it('handleConfigChange ignores NaN values', async () => {
    const { result } = renderHook(() => useCrazyEightsGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    act(() => {
      result.current.handleConfigChange('pointLimit', 'abc');
    });

    expect(result.current.crazyEightsConfig.pointLimit).toBe(200);
  });

  it('clears selection on success', async () => {
    const { result } = renderHook(() => useCrazyEightsGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    act(() => {
      result.current.toggleCard(0);
    });
    expect(result.current.selectedCardIndices).toEqual([0]);

    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handlePlay();
    });

    await waitFor(() => expect(result.current.selectedCardIndices).toEqual([]));
  });
});
