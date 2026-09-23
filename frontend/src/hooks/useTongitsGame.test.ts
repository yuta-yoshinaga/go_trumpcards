import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { createElement } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { tongitsApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import type { TongitsResponse } from '../types/card';
import { useTongitsGame } from './useTongitsGame';

vi.mock('../api/gameApi', () => ({
  tongitsApi: { exec: vi.fn() },
}));

const mockExec = vi.mocked(tongitsApi.exec);

function createWrapper() {
  const queryClient = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  return ({ children }: { children: ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children);
}

const defaultState: TongitsResponse = {
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 5,
      cards: [{ design: 'SPADE', value: 8 }],
      melds: [],
      roundScore: 0,
      cumulativeScore: 0,
    },
    { id: 1, isHuman: false, cardCount: 5, cards: [], melds: [], roundScore: 0, cumulativeScore: 0 },
  ],
  phase: 0,
  roundNumber: 1,
  currentPlayerIdx: 0,
  discardTop: { design: 'HEART', value: 7 },
  drawPileCount: 41,
  gameEndFlag: false,
  winnerIdx: -1,
  isTongits: false,
  remainingPoints: -1,
  message: '',
  config: { cpuDifficulty: 1, pointLimit: 50 },
};

beforeEach(() => {
  mockExec.mockResolvedValue(defaultState);
});

describe('useTongitsGame', () => {
  it('calls reset on mount with default config', async () => {
    renderHook(() => useTongitsGame(), { wrapper: createWrapper() });
    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('reset', undefined, {
        cpuDifficulty: 1,
        pointLimit: 50,
      }),
    );
  });

  it('returns initial state after mount', async () => {
    const { result } = renderHook(() => useTongitsGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).toEqual(defaultState));
  });

  it('handleDrawStock dispatches drawstock command', async () => {
    const { result } = renderHook(() => useTongitsGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handleDrawStock();
    });

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('drawstock'));
  });

  it('handleDrawDiscard dispatches drawdiscard command', async () => {
    const { result } = renderHook(() => useTongitsGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handleDrawDiscard();
    });

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('drawdiscard'));
  });

  it('handleDiscard dispatches discard with single selected card', async () => {
    const { result } = renderHook(() => useTongitsGame(), { wrapper: createWrapper() });
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
    const { result } = renderHook(() => useTongitsGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    act(() => {
      result.current.handleDiscard();
    });

    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('handleChallenge dispatches challenge with agreement', async () => {
    const { result } = renderHook(() => useTongitsGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    act(() => {
      result.current.handleChallenge();
    });

    await waitFor(() =>
      expect(mockExec).toHaveBeenCalledWith('challenge', undefined, undefined, undefined, undefined, undefined, [
        true,
        true,
      ]),
    );
  });

  it('handleMeld does nothing with fewer than three selected cards', async () => {
    const { result } = renderHook(() => useTongitsGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    act(() => {
      result.current.handleMeld();
    });

    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('handleNextRound dispatches nextround command', async () => {
    const { result } = renderHook(() => useTongitsGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handleNextRound();
    });

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('nextround'));
  });

  it('handleConfigChange updates pointLimit', async () => {
    const { result } = renderHook(() => useTongitsGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    act(() => {
      result.current.handleConfigChange('pointLimit', '100');
    });

    expect(result.current.tongitsConfig.pointLimit).toBe(100);
  });

  it('clears selection on success', async () => {
    const { result } = renderHook(() => useTongitsGame(), { wrapper: createWrapper() });
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
