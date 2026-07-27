import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { createElement } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { machiavelliApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import type { MachiavelliResponse } from '../types/card';
import { useMachiavelliGame } from './useMachiavelliGame';

vi.mock('../api/gameApi', () => ({
  machiavelliApi: { exec: vi.fn() },
}));

const mockExec = vi.mocked(machiavelliApi.exec);

function createWrapper() {
  const queryClient = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  return ({ children }: { children: ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children);
}

const defaultState: MachiavelliResponse = {
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 13,
      cards: [{ design: 'SPADE', value: 8 }],
      roundScore: 0,
      cumulativeScore: 0,
      deadwood: 0,
    },
    {
      id: 1,
      isHuman: false,
      cardCount: 13,
      cards: [],
      roundScore: 0,
      cumulativeScore: 0,
      deadwood: 0,
    },
  ],
  table: [],
  phase: 0,
  roundNumber: 1,
  targetRounds: 3,
  currentPlayerIdx: 0,
  dealerIdx: 0,
  drawPileCount: 40,
  gameEndFlag: false,
  winnerIdx: -1,
  roundWinnerIdx: -1,
  message: '',
  config: { playerCount: 4, cpuDifficulty: 1, targetRounds: 3 },
};

beforeEach(() => {
  mockExec.mockResolvedValue(defaultState);
});

describe('useMachiavelliGame', () => {
  it('calls reset on mount with default config', async () => {
    renderHook(() => useMachiavelliGame(), { wrapper: createWrapper() });
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, {
        playerCount: 4,
        cpuDifficulty: 1,
        targetRounds: 3,
      }),
    );
  });

  it('returns initial state after mount', async () => {
    const { result } = renderHook(() => useMachiavelliGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).toEqual(defaultState));
  });

  it('handleDraw dispatches draw command', async () => {
    const { result } = renderHook(() => useMachiavelliGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handleDraw();
    });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('draw'));
  });

  it('handleNewMeld dispatches newmeld with the selected indices', async () => {
    const { result } = renderHook(() => useMachiavelliGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    act(() => {
      result.current.toggleCard(0);
      result.current.toggleCard(1);
      result.current.toggleCard(2);
    });
    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handleNewMeld();
    });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('newmeld', { handIndices: [0, 1, 2] }));
  });

  it('handleNewMeld does nothing with fewer than 3 cards', async () => {
    const { result } = renderHook(() => useMachiavelliGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    act(() => {
      result.current.toggleCard(0);
      result.current.toggleCard(1);
    });
    mockExec.mockClear();
    act(() => {
      result.current.handleNewMeld();
    });
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('handleLayoff dispatches layoff with meld index and the single selected card', async () => {
    const { result } = renderHook(() => useMachiavelliGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    act(() => {
      result.current.toggleCard(2);
    });
    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handleLayoff(1);
    });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('layoff', { meldIdx: 1, handIndex: 2 }));
  });

  it('handleLayoff does nothing when not exactly one card selected', async () => {
    const { result } = renderHook(() => useMachiavelliGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    act(() => {
      result.current.handleLayoff(0);
    });
    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('handleNextRound dispatches nextround command', async () => {
    const { result } = renderHook(() => useMachiavelliGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handleNextRound();
    });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
  });

  it('handleConfigChange updates config with a valid number', async () => {
    const { result } = renderHook(() => useMachiavelliGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    act(() => {
      result.current.handleConfigChange('playerCount', '5');
    });
    expect(result.current.machiavelliConfig.playerCount).toBe(5);
  });

  it('handleConfigChange ignores NaN values', async () => {
    const { result } = renderHook(() => useMachiavelliGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    act(() => {
      result.current.handleConfigChange('playerCount', 'abc');
    });
    expect(result.current.machiavelliConfig.playerCount).toBe(4);
  });

  it('clears selection on success', async () => {
    const { result } = renderHook(() => useMachiavelliGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    act(() => {
      result.current.toggleCard(0);
    });
    expect(result.current.selectedCardIndices).toEqual([0]);

    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handleDraw();
    });
    await waitFor(() => expect(result.current.selectedCardIndices).toEqual([]));
  });
});
