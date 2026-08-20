import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { act, renderHook, waitFor } from '@testing-library/react';
import type { ReactNode } from 'react';
import { createElement } from 'react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { daifugoApi } from '../api/gameApi';
import type { DaifugoResponse } from '../types/card';
import * as gameReplay from './gameReplay';
import { useDaifugoGame } from './useDaifugoGame';

vi.mock('../api/gameApi', () => ({
  daifugoApi: { exec: vi.fn() },
}));

const mockExec = vi.mocked(daifugoApi.exec);

function createWrapper() {
  const queryClient = new QueryClient({ defaultOptions: { mutations: { retry: false } } });
  return ({ children }: { children: ReactNode }) =>
    createElement(QueryClientProvider, { client: queryClient }, children);
}

const baseState: DaifugoResponse = {
  players: [
    { id: 0, isHuman: true, cardCount: 5, cards: [], rank: 0, isFinished: false },
    { id: 1, isHuman: false, cardCount: 5, cards: [], rank: 0, isFinished: false },
  ],
  currentTurn: 0,
  tableCards: [],
  lastPlayPlayerIdx: -1,
  gameEndFlag: false,
  revolutionActive: false,
  elevenBackActive: false,
  suitLocked: false,
  lockedSuit: '',
  tableIsSequence: false,
  config: {} as DaifugoResponse['config'],
  exchangeActions: [],
  cpuActions: [
    {
      playerIdx: 1,
      playedCards: [{ design: 'SPADE', value: 3 }],
    },
  ],
  humanAction: null,
  message: '',
  pendingAction: 'none',
  pendingActionTarget: -1,
  reverseDirection: false,
  numberLocked: false,
  sequenceLocked: false,
  sortMode: 0,
  playableCardIndices: null,
};

describe('useDaifugoGame onSuccess replay skip', () => {
  let runReplaySpy: ReturnType<typeof vi.spyOn>;

  beforeEach(() => {
    vi.restoreAllMocks();
    runReplaySpy = vi.spyOn(gameReplay, 'runReplay').mockResolvedValue(undefined);
    mockExec.mockResolvedValue(baseState);
  });

  it('calls runReplay on first success (no previous actions)', async () => {
    renderHook(() => useDaifugoGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(runReplaySpy).toHaveBeenCalledTimes(1));
  });

  it('skips replay when cpuActions unchanged', async () => {
    const { result } = renderHook(() => useDaifugoGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(runReplaySpy).toHaveBeenCalledTimes(1));

    // Second call with same cpuActions
    runReplaySpy.mockClear();
    mockExec.mockResolvedValue(baseState);
    act(() => {
      result.current.exec('sort');
    });
    await waitFor(() => expect(result.current.state).toBeDefined());
    expect(runReplaySpy).not.toHaveBeenCalled();
  });

  it('runs replay when cpuActions change', async () => {
    const { result } = renderHook(() => useDaifugoGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(runReplaySpy).toHaveBeenCalledTimes(1));

    const newState: DaifugoResponse = {
      ...baseState,
      cpuActions: [
        {
          playerIdx: 1,
          playedCards: [{ design: 'HEART', value: 5 }],
        },
      ],
    };
    runReplaySpy.mockClear();
    mockExec.mockResolvedValue(newState);
    act(() => {
      result.current.exec('play', [0]);
    });
    await waitFor(() => expect(runReplaySpy).toHaveBeenCalledTimes(1));
  });

  it('sets displayState even when replay is skipped', async () => {
    const { result } = renderHook(() => useDaifugoGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(runReplaySpy).toHaveBeenCalledTimes(1));

    // Second call with same cpuActions → replay skipped, but state should update
    const updatedState: DaifugoResponse = { ...baseState, sortMode: 1 };
    runReplaySpy.mockClear();
    mockExec.mockResolvedValue(updatedState);
    act(() => {
      result.current.exec('sort');
    });
    await waitFor(() => expect(result.current.state?.sortMode).toBe(1));
    expect(runReplaySpy).not.toHaveBeenCalled();
  });

  it('runs replay when previous cpuActions was empty and new has actions', async () => {
    const emptyActionsState: DaifugoResponse = { ...baseState, cpuActions: [] };
    mockExec.mockResolvedValue(emptyActionsState);
    const { result } = renderHook(() => useDaifugoGame(), { wrapper: createWrapper() });
    await waitFor(() => expect(runReplaySpy).toHaveBeenCalledTimes(1));

    // Now with actions
    runReplaySpy.mockClear();
    mockExec.mockResolvedValue(baseState);
    act(() => {
      result.current.exec('play', [0]);
    });
    await waitFor(() => expect(runReplaySpy).toHaveBeenCalledTimes(1));
  });
});
