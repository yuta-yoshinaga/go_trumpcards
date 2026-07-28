import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { createElement } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { ginrummyApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import type { GinRummyResponse } from '../types/card';
import { useGinRummyGame } from './useGinRummyGame';

vi.mock('../api/gameApi', () => ({
  ginrummyApi: { exec: vi.fn() },
}));

const mockExec = vi.mocked(ginrummyApi.exec);

function createWrapper() {
  const queryClient = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  return ({ children }: { children: ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children);
}

const defaultState: GinRummyResponse = {
  players: [
    { id: 0, isHuman: true, cardCount: 10, cards: [{ design: 'SPADE', value: 8 }], roundScore: 0, cumulativeScore: 0 },
    { id: 1, isHuman: false, cardCount: 10, cards: [], roundScore: 0, cumulativeScore: 0 },
  ],
  phase: 0,
  roundNumber: 1,
  currentPlayerIdx: 0,
  discardTop: { design: 'HEART', value: 7 },
  drawPileCount: 30,
  gameEndFlag: false,
  winnerIdx: -1,
  knockerIdx: -1,
  knockerMelds: [],
  knockerDeadwood: [],
  isGin: false,
  message: '',
  config: { cpuDifficulty: 1, pointLimit: 100 },
};

beforeEach(() => {
  mockExec.mockResolvedValue(defaultState);
});

describe('useGinRummyGame', () => {
  it('calls reset on mount with default config', async () => {
    renderHook(() => useGinRummyGame(), { wrapper: createWrapper() });
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, {
        cpuDifficulty: 1,
        pointLimit: 100,
      }),
    );
  });

  it('returns initial state after mount', async () => {
    const { result } = renderHook(() => useGinRummyGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).toEqual(defaultState));
  });

  it('handleDrawStock dispatches drawstock command', async () => {
    const { result } = renderHook(() => useGinRummyGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handleDrawStock();
    });

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('drawstock'));
  });

  it('handleDrawDiscard dispatches drawdiscard command', async () => {
    const { result } = renderHook(() => useGinRummyGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handleDrawDiscard();
    });

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('drawdiscard'));
  });

  it('handleDiscard dispatches discard with single selected card', async () => {
    const { result } = renderHook(() => useGinRummyGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    act(() => {
      result.current.toggleCard(2);
    });

    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handleDiscard();
    });

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('discard', 2));
  });

  it('handleDiscard does nothing when no card selected', async () => {
    const { result } = renderHook(() => useGinRummyGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    act(() => {
      result.current.handleDiscard();
    });

    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('handleDiscard does nothing when multiple cards selected', async () => {
    const { result } = renderHook(() => useGinRummyGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    act(() => {
      result.current.toggleCard(0);
      result.current.toggleCard(1);
    });

    mockExec.mockClear();
    act(() => {
      result.current.handleDiscard();
    });

    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('handleKnock dispatches knock with single selected card', async () => {
    const { result } = renderHook(() => useGinRummyGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    act(() => {
      result.current.toggleCard(3);
    });

    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handleKnock();
    });

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('knock', 3));
  });

  it('handleKnock does nothing when no card selected', async () => {
    const { result } = renderHook(() => useGinRummyGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    act(() => {
      result.current.handleKnock();
    });

    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('handleKnock does nothing when multiple cards selected', async () => {
    const { result } = renderHook(() => useGinRummyGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    act(() => {
      result.current.toggleCard(0);
      result.current.toggleCard(1);
    });

    mockExec.mockClear();
    act(() => {
      result.current.handleKnock();
    });

    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('handleLayoff dispatches layoff with selected card indices', async () => {
    const { result } = renderHook(() => useGinRummyGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    act(() => {
      result.current.toggleCard(0);
      result.current.toggleCard(2);
    });

    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handleLayoff();
    });

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('layoff', undefined, undefined, [0, 2]));
  });

  it('handleSkipLayoff dispatches layoff with empty array', async () => {
    const { result } = renderHook(() => useGinRummyGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handleSkipLayoff();
    });

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('layoff', undefined, undefined, []));
  });

  it('handleNextRound dispatches nextround command', async () => {
    const { result } = renderHook(() => useGinRummyGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handleNextRound();
    });

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
  });

  it('handleConfigChange updates config with valid number', async () => {
    const { result } = renderHook(() => useGinRummyGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    act(() => {
      result.current.handleConfigChange('pointLimit', '150');
    });

    expect(result.current.ginRummyConfig.pointLimit).toBe(150);
  });

  it('handleConfigChange ignores NaN values', async () => {
    const { result } = renderHook(() => useGinRummyGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    act(() => {
      result.current.handleConfigChange('pointLimit', 'abc');
    });

    expect(result.current.ginRummyConfig.pointLimit).toBe(100);
  });

  it('clears selection on success', async () => {
    const { result } = renderHook(() => useGinRummyGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    act(() => {
      result.current.toggleCard(0);
    });
    expect(result.current.selectedCardIndices).toEqual([0]);

    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handleDrawStock();
    });

    await waitFor(() => expect(result.current.selectedCardIndices).toEqual([]));
  });
});
