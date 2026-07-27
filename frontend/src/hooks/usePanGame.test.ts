import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { createElement } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { panApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import type { PanResponse } from '../types/card';
import { usePanGame } from './usePanGame';

vi.mock('../api/gameApi', () => ({
  panApi: { exec: vi.fn() },
}));

const mockExec = vi.mocked(panApi.exec);

function createWrapper() {
  const queryClient = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  return ({ children }: { children: ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children);
}

const defaultState: PanResponse = {
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 11,
      cards: [{ design: 'SPADE', value: 3 }],
      laidMelds: [],
      meldedCount: 0,
      chips: 50,
      handPoints: 0,
      roundScore: 0,
      cumulativeScore: 0,
    },
    {
      id: 1,
      isHuman: false,
      cardCount: 10,
      cards: [],
      laidMelds: [],
      meldedCount: 0,
      chips: 50,
      handPoints: 0,
      roundScore: 0,
      cumulativeScore: 0,
    },
  ],
  phase: 0,
  roundNumber: 1,
  targetRounds: 3,
  currentPlayerIdx: 0,
  dealerIdx: 0,
  discardTop: { design: 'HEART', value: 7 },
  drawPileCount: 250,
  deckSize: 320,
  winMeldCount: 11,
  gameEndFlag: false,
  winnerIdx: -1,
  panDeclarerIdx: -1,
  message: '',
  config: { playerCount: 4, cpuDifficulty: 1, targetRounds: 3 },
};

beforeEach(() => {
  mockExec.mockResolvedValue(defaultState);
});

describe('usePanGame', () => {
  it('calls reset on mount with default config', async () => {
    renderHook(() => usePanGame(), { wrapper: createWrapper() });
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, {
        playerCount: 4,
        cpuDifficulty: 1,
        targetRounds: 3,
      }),
    );
  });

  it('returns initial state after mount', async () => {
    const { result } = renderHook(() => usePanGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).toEqual(defaultState));
  });

  it('handleDrawStock dispatches drawstock command', async () => {
    const { result } = renderHook(() => usePanGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());
    mockExec.mockClear();
    act(() => {
      result.current.handleDrawStock();
    });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('drawstock'));
  });

  it('handleDrawDiscard dispatches drawdiscard command', async () => {
    const { result } = renderHook(() => usePanGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());
    mockExec.mockClear();
    act(() => {
      result.current.handleDrawDiscard();
    });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('drawdiscard'));
  });

  it('handleMeld dispatches meld with the selected indices', async () => {
    const { result } = renderHook(() => usePanGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());
    act(() => {
      result.current.toggleCard(0);
      result.current.toggleCard(1);
      result.current.toggleCard(2);
    });
    mockExec.mockClear();
    act(() => {
      result.current.handleMeld();
    });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('meld', { cardIndices: [0, 1, 2] }));
  });

  it('handleMeld does nothing with fewer than 3 cards', async () => {
    const { result } = renderHook(() => usePanGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());
    act(() => {
      result.current.toggleCard(0);
      result.current.toggleCard(1);
    });
    mockExec.mockClear();
    act(() => {
      result.current.handleMeld();
    });
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('handleLayoff dispatches layoff with the target meld and selected card', async () => {
    const { result } = renderHook(() => usePanGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());
    act(() => {
      result.current.toggleCard(4);
    });
    mockExec.mockClear();
    act(() => {
      result.current.handleLayoff(1, 0);
    });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('layoff', { meldOwner: 1, meldIdx: 0, cardIndex: 4 }));
  });

  it('handleLayoff does nothing when not exactly 1 card selected', async () => {
    const { result } = renderHook(() => usePanGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());
    mockExec.mockClear();
    act(() => {
      result.current.handleLayoff(1, 0);
    });
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('handleDiscard dispatches discard with the single selected card', async () => {
    const { result } = renderHook(() => usePanGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());
    act(() => {
      result.current.toggleCard(2);
    });
    mockExec.mockClear();
    act(() => {
      result.current.handleDiscard();
    });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('discard', { cardIndex: 2 }));
  });

  it('handleDiscard does nothing when no card selected', async () => {
    const { result } = renderHook(() => usePanGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());
    mockExec.mockClear();
    act(() => {
      result.current.handleDiscard();
    });
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('handleNextRound dispatches nextround command', async () => {
    const { result } = renderHook(() => usePanGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());
    mockExec.mockClear();
    act(() => {
      result.current.handleNextRound();
    });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
  });

  it('handleConfigChange updates config with a valid number', async () => {
    const { result } = renderHook(() => usePanGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());
    act(() => {
      result.current.handleConfigChange('playerCount', '5');
    });
    expect(result.current.panConfig.playerCount).toBe(5);
  });

  it('handleConfigChange ignores NaN values', async () => {
    const { result } = renderHook(() => usePanGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());
    act(() => {
      result.current.handleConfigChange('playerCount', 'abc');
    });
    expect(result.current.panConfig.playerCount).toBe(4);
  });

  it('clears selection on success', async () => {
    const { result } = renderHook(() => usePanGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());
    act(() => {
      result.current.toggleCard(0);
    });
    expect(result.current.selectedCardIndices).toEqual([0]);
    act(() => {
      result.current.handleDrawStock();
    });
    await waitFor(() => expect(result.current.selectedCardIndices).toEqual([]));
  });
});
