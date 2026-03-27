import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { createElement } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { oldmaidApi } from '../api/gameApi';
import type { OldMaidResponse } from '../types/card';
import * as gameReplay from './gameReplay';
import { useOldMaidGame } from './useOldMaidGame';

vi.mock('../api/gameApi', () => ({
  oldmaidApi: { exec: vi.fn() },
}));

const mockExec = vi.mocked(oldmaidApi.exec);

function createWrapper() {
  const queryClient = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  return ({ children }: { children: ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children);
}

const baseState: OldMaidResponse = {
  players: [
    { id: 0, isHuman: true, cardCount: 5, cards: [], isFinished: false },
    { id: 1, isHuman: false, cardCount: 5, cards: [], isFinished: false },
  ],
  currentTurn: 0,
  hasDrawn: false,
  lastDrawPlayerIdx: -1,
  lastDrawFromIdx: -1,
  lastDrawCard: null,
  lastDiscardedPairs: 0,
  lastDiscardedCards: [],
  cpuActions: [
    {
      drawPlayerIdx: 1,
      drawFromIdx: 0,
      drawnCard: { design: 'SPADE', value: 3 },
      discardedPairs: 0,
    },
  ],
  humanAction: null,
  drawHistory: [],
  nextDrawTargetIdx: 0,
  gameEndFlag: false,
  message: '',
  cpuHighlightedCardIdx: -1,
  removedCard: null,
  mode: 0,
};

describe('useOldMaidGame onSuccess replay skip', () => {
  let runReplaySpy: ReturnType<typeof vi.spyOn>;

  beforeEach(() => {
    vi.restoreAllMocks();
    runReplaySpy = vi.spyOn(gameReplay, 'runReplay').mockResolvedValue(undefined);
    mockExec.mockResolvedValue(baseState);
  });

  it('calls runReplay on auto-start', async () => {
    renderHook(() => useOldMaidGame(), { wrapper: createWrapper() });
    // Auto-start triggers reset on mount
    await waitFor(() => expect(runReplaySpy).toHaveBeenCalledTimes(1));
  });

  it('skips replay when cpuActions unchanged', async () => {
    const { result } = renderHook(() => useOldMaidGame(), { wrapper: createWrapper() });
    // Wait for auto-start
    await waitFor(() => expect(runReplaySpy).toHaveBeenCalledTimes(1));

    runReplaySpy.mockClear();
    mockExec.mockResolvedValue(baseState);
    act(() => {
      result.current.handleReorder([1, 0, 2, 3, 4]);
    });
    await waitFor(() => expect(result.current.displayState).toBeDefined());
    expect(runReplaySpy).not.toHaveBeenCalled();
  });

  it('runs replay when cpuActions change', async () => {
    const { result } = renderHook(() => useOldMaidGame(), { wrapper: createWrapper() });
    // Wait for auto-start
    await waitFor(() => expect(runReplaySpy).toHaveBeenCalledTimes(1));

    const newState: OldMaidResponse = {
      ...baseState,
      cpuActions: [
        {
          drawPlayerIdx: 1,
          drawFromIdx: 0,
          drawnCard: { design: 'HEART', value: 5 },
          discardedPairs: 1,
        },
      ],
    };
    runReplaySpy.mockClear();
    mockExec.mockResolvedValue(newState);
    act(() => {
      result.current.gameExec('draw', 0);
    });
    await waitFor(() => expect(runReplaySpy).toHaveBeenCalledTimes(1));
  });

  it('sets displayState even when replay is skipped', async () => {
    const { result } = renderHook(() => useOldMaidGame(), { wrapper: createWrapper() });
    // Wait for auto-start
    await waitFor(() => expect(runReplaySpy).toHaveBeenCalledTimes(1));

    const updatedState: OldMaidResponse = { ...baseState, currentTurn: 1 };
    runReplaySpy.mockClear();
    mockExec.mockResolvedValue(updatedState);
    act(() => {
      result.current.handleReorder([1, 0, 2, 3, 4]);
    });
    await waitFor(() => expect(result.current.displayState?.currentTurn).toBe(1));
    expect(runReplaySpy).not.toHaveBeenCalled();
  });
});
