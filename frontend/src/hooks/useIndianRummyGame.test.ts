import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { createElement } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { indianRummyApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import type { IndianRummyResponse } from '../types/card';
import { useIndianRummyGame } from './useIndianRummyGame';

vi.mock('../api/gameApi', () => ({
  indianRummyApi: { exec: vi.fn() },
}));

const mockExec = vi.mocked(indianRummyApi.exec);

function createWrapper() {
  const queryClient = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  return ({ children }: { children: ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children);
}

const defaultState: IndianRummyResponse = {
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 13,
      cards: [{ design: 'SPADE', value: 8 }],
      roundScore: 0,
      cumulativeScore: 0,
      deadwood: 0,
      hasPureSequence: false,
    },
    {
      id: 1,
      isHuman: false,
      cardCount: 13,
      cards: [],
      roundScore: 0,
      cumulativeScore: 0,
      deadwood: 0,
      hasPureSequence: false,
    },
  ],
  phase: 0,
  roundNumber: 1,
  targetRounds: 3,
  currentPlayerIdx: 0,
  dealerIdx: 0,
  discardTop: { design: 'HEART', value: 7 },
  drawPileCount: 40,
  wildJoker: { design: 'CLOVER', value: 5 },
  wildRank: 5,
  gameEndFlag: false,
  winnerIdx: -1,
  declarerIdx: -1,
  declarationValid: false,
  message: '',
  config: { playerCount: 4, cpuDifficulty: 1, targetRounds: 3 },
};

beforeEach(() => {
  mockExec.mockResolvedValue(defaultState);
});

describe('useIndianRummyGame', () => {
  it('calls reset on mount with default config', async () => {
    renderHook(() => useIndianRummyGame(), { wrapper: createWrapper() });
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, {
        playerCount: 4,
        cpuDifficulty: 1,
        targetRounds: 3,
      }),
    );
  });

  it('returns initial state after mount', async () => {
    const { result } = renderHook(() => useIndianRummyGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).toEqual(defaultState));
  });

  it('handleDrawStock dispatches drawstock command', async () => {
    const { result } = renderHook(() => useIndianRummyGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handleDrawStock();
    });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('drawstock'));
  });

  it('handleDrawDiscard dispatches drawdiscard command', async () => {
    const { result } = renderHook(() => useIndianRummyGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handleDrawDiscard();
    });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('drawdiscard'));
  });

  it('handleDiscard dispatches discard with the single selected card', async () => {
    const { result } = renderHook(() => useIndianRummyGame(), { wrapper: createWrapper() });
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
    const { result } = renderHook(() => useIndianRummyGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    act(() => {
      result.current.handleDiscard();
    });
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('handleDiscard does nothing when multiple cards selected', async () => {
    const { result } = renderHook(() => useIndianRummyGame(), { wrapper: createWrapper() });
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

  it('handleDeclare dispatches declare with the single selected card', async () => {
    const { result } = renderHook(() => useIndianRummyGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    act(() => {
      result.current.toggleCard(3);
    });
    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handleDeclare();
    });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('declare', 3));
  });

  it('handleDeclare does nothing when no card selected', async () => {
    const { result } = renderHook(() => useIndianRummyGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    act(() => {
      result.current.handleDeclare();
    });
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('handleNextRound dispatches nextround command', async () => {
    const { result } = renderHook(() => useIndianRummyGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handleNextRound();
    });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
  });

  it('handleConfigChange updates config with a valid number', async () => {
    const { result } = renderHook(() => useIndianRummyGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    act(() => {
      result.current.handleConfigChange('playerCount', '3');
    });
    expect(result.current.indianRummyConfig.playerCount).toBe(3);
  });

  it('handleConfigChange ignores NaN values', async () => {
    const { result } = renderHook(() => useIndianRummyGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    act(() => {
      result.current.handleConfigChange('playerCount', 'abc');
    });
    expect(result.current.indianRummyConfig.playerCount).toBe(4);
  });

  it('clears selection on success', async () => {
    const { result } = renderHook(() => useIndianRummyGame(), { wrapper: createWrapper() });
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
