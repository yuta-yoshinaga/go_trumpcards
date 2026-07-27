import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { createElement } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { goFishApi } from '../api/gameApi';
import { flushPendingDispatch } from '../test/flushPendingDispatch';
import type { GoFishResponse } from '../types/card';
import { GoFishPhase } from '../types/phases';
import { useGoFishGame } from './useGoFishGame';

vi.mock('../api/gameApi', () => ({
  goFishApi: { exec: vi.fn() },
}));

const mockExec = vi.mocked(goFishApi.exec);

function createWrapper() {
  const queryClient = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  return ({ children }: { children: ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children);
}

const defaultState: GoFishResponse = {
  players: [
    { id: 0, isHuman: true, cardCount: 5, cards: [{ design: 'SPADE', value: 7 }], bookCount: 0, books: [] },
    { id: 1, isHuman: false, cardCount: 5, cards: [], bookCount: 0, books: [] },
    { id: 2, isHuman: false, cardCount: 5, cards: [], bookCount: 0, books: [] },
    { id: 3, isHuman: false, cardCount: 5, cards: [], bookCount: 0, books: [] },
  ],
  phase: GoFishPhase.PLAY,
  currentTurn: 0,
  gameEndFlag: false,
  winnerIdx: -1,
  turnNumber: 1,
  deckRemaining: 32,
  lastAsk: null,
  cpuActions: [],
  humanAction: null,
  message: '',
  config: { cpuDifficulty: 1 },
};

beforeEach(() => {
  mockExec.mockResolvedValue(defaultState);
});

describe('useGoFishGame', () => {
  it('calls reset on mount with default config', async () => {
    renderHook(() => useGoFishGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('reset', undefined, undefined, { cpuDifficulty: 1 }));
  });

  it('returns initial state after mount', async () => {
    const { result } = renderHook(() => useGoFishGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).toEqual(defaultState));
  });

  it('handleSelectTarget sets selected target', async () => {
    const { result } = renderHook(() => useGoFishGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    act(() => {
      result.current.handleSelectTarget(2);
    });

    expect(result.current.selectedTarget).toBe(2);
  });

  it('handleSelectTarget toggles off when same target clicked again', async () => {
    const { result } = renderHook(() => useGoFishGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    act(() => {
      result.current.handleSelectTarget(2);
    });
    expect(result.current.selectedTarget).toBe(2);

    act(() => {
      result.current.handleSelectTarget(2);
    });
    expect(result.current.selectedTarget).toBeNull();
  });

  it('handleSelectRank sets selected rank', async () => {
    const { result } = renderHook(() => useGoFishGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    act(() => {
      result.current.handleSelectRank(7);
    });

    expect(result.current.selectedRank).toBe(7);
  });

  it('handleSelectRank toggles off when same rank clicked again', async () => {
    const { result } = renderHook(() => useGoFishGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    act(() => {
      result.current.handleSelectRank(7);
    });
    expect(result.current.selectedRank).toBe(7);

    act(() => {
      result.current.handleSelectRank(7);
    });
    expect(result.current.selectedRank).toBeNull();
  });

  it('handleAsk does nothing when target not selected', async () => {
    const { result } = renderHook(() => useGoFishGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    act(() => {
      result.current.handleSelectRank(7);
    });
    mockExec.mockClear();
    act(() => {
      result.current.handleAsk();
    });

    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('handleAsk does nothing when rank not selected', async () => {
    const { result } = renderHook(() => useGoFishGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    act(() => {
      result.current.handleSelectTarget(1);
    });
    mockExec.mockClear();
    act(() => {
      result.current.handleAsk();
    });

    await flushPendingDispatch();
    expect(mockExec).not.toHaveBeenCalled();
  });

  it('handleAsk dispatches ask command with target and rank', async () => {
    const { result } = renderHook(() => useGoFishGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    act(() => {
      result.current.handleSelectTarget(2);
      result.current.handleSelectRank(7);
    });

    mockExec.mockClear();
    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handleAsk();
    });

    await waitFor(() => expect(mockExec).toHaveBeenCalledWith('ask', 2, 7));
  });

  it('clears selections on successful ask', async () => {
    const { result } = renderHook(() => useGoFishGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    act(() => {
      result.current.handleSelectTarget(1);
      result.current.handleSelectRank(5);
    });
    expect(result.current.selectedTarget).toBe(1);
    expect(result.current.selectedRank).toBe(5);

    mockExec.mockResolvedValue(defaultState);
    act(() => {
      result.current.handleAsk();
    });

    await waitFor(() => {
      expect(result.current.selectedTarget).toBeNull();
      expect(result.current.selectedRank).toBeNull();
    });
  });

  it('handleConfigChange updates cpuDifficulty', async () => {
    const { result } = renderHook(() => useGoFishGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(result.current.state).not.toBeNull());

    act(() => {
      result.current.handleConfigChange('cpuDifficulty', '2');
    });

    expect(result.current.goFishConfig.cpuDifficulty).toBe(2);
  });
});
